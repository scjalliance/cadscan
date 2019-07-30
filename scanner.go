package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// A token represents the right to perform work
type token struct{}

type task struct {
	path   string
	result chan File
}

// A Scanner is responsible for scanning file systems for DWG files and
// assessing their versions.
//
// Scanner records its results to a ScanModel.
type Scanner struct {
	model  *ScanModel
	tokens chan token

	mutex   sync.Mutex
	cancel  context.CancelFunc
	stopped <-chan struct{}
}

// NewScanner returns a scanner that will write its output to the given model.
func NewScanner(model *ScanModel, workers int) *Scanner {
	s := &Scanner{
		model:  model,
		tokens: make(chan token, workers),
	}
	for i := 0; i < workers; i++ {
		s.tokens <- token{}
	}
	return s
}

// Scan causes the scanner to start scanning the given directory.
func (s *Scanner) Scan(dir string, init, finished func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stop()

	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})
	s.cancel, s.stopped = cancel, stopped

	go s.scan(ctx, stopped, dir, init, finished)
}

// Stop cancels any scan that may be in-progress.
func (s *Scanner) Stop() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.stop()
}

// stop cancels any scan that may be in-progress. The caller must hold a lock
// on s.mutex for the duration of this call.
func (s *Scanner) stop() bool {
	if s.cancel == nil {
		return false
	}

	var alreadyStopped bool
	select {
	case <-s.stopped:
		alreadyStopped = true
	default:
		s.cancel()
		<-s.stopped
	}

	s.cancel = nil
	s.stopped = nil

	if alreadyStopped {
		return false
	}

	return true
}

func (s *Scanner) scan(ctx context.Context, done chan<- struct{}, dir string, init, finished func()) {
	defer close(done)
	if finished != nil {
		defer finished()
	}

	if init != nil {
		init()
	}

	s.model.Clear()

	queue := make(chan string, 128)      // Ordered paths to be scanned
	results := make(chan chan File, 128) // Ordered results

	// Phase 1: Harvest paths from the file system
	go func() {
		defer close(queue)
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err != nil || info == nil {
				return filepath.SkipDir
			}

			name := info.Name()
			if info.IsDir() {
				if shouldExclude(name) {
					return filepath.SkipDir
				}
				return nil
			}

			if isDrawingFile(name) {
				queue <- path
			}

			return nil
		})
	}()

	// Phase 2: Spawn workers for each path
	go func() {
		defer close(results)
		for path := range queue {
			select {
			case <-s.tokens:
			case <-ctx.Done():
				for range queue {
					// Drain the channel and exit when cancelled
				}
				return
			}

			result := make(chan File, 1)

			go func(path string, result chan<- File) {
				defer close(result)
				//fmt.Printf("Scanning %s\n", path)
				if version, err := ReadDrawingVersion(path); err == nil {
					result <- File{Path: path, Version: version}
				}
				s.tokens <- token{}
			}(path, result)

			results <- result
		}
	}()

	// Phase 3: Collect results in order and append them to the table
	t := time.NewTicker(time.Millisecond * 200)
	defer t.Stop()

	var drained bool
	var batch []File
	for !drained {
		select {
		case result, ok := <-results:
			if !ok {
				drained = true
				s.model.Append(batch...)
				break
			}
			if file, ok := <-result; ok {
				batch = append(batch, file)
			}
		case <-t.C:
			s.model.Append(batch...)
			batch = batch[:0]
		}
	}
}

func isDrawingFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".dwg")
}
