package main

import (
	"errors"
	"io"
	"os"
)

// ErrInvalidDrawing is returned when a file does not posses a drawing header.
var ErrInvalidDrawing = errors.New("not a drawing file")

// DrawingVersion is a drawing version string from a drawing header.
//
// The string should match one of values specified in this document:
// https://knowledge.autodesk.com/support/autocad/learn-explore/caas/sfdcarticles/sfdcarticles/drawing-version-codes-for-autocad.html
type DrawingVersion string

// ReadDrawingVersion attempts to open the file with the given name and
// return its drawing format header.
func ReadDrawingVersion(name string) (DrawingVersion, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var header [6]byte

	n, err := f.Read(header[:])
	if err != nil && err != io.EOF {
		return "", err
	}
	if n < 6 {
		return "", ErrInvalidDrawing
	}

	h := header[:]
	switch {
	case h[0] == 'A' && h[1] == 'C':
		switch {
		case h[2] == '1' && h[3] == '.' && h[4] == '2':
			return "AC1.2", nil
		case h[2] == '1' && h[3] == '.' && h[4] == '4':
			return "AC1.4", nil
		default:
			return DrawingVersion(h), nil
		}
	case h[0] == 'M' && h[1] == 'C' && h[2] == '0' && h[3] == '.' && h[4] == '0':
		return "MC0.0", nil
	default:
		return "", ErrInvalidDrawing
	}
}

// String returns a string representation of the drawing version.
func (v DrawingVersion) String() string {
	return string(v)
}

// Release returns an ordered integer representing the drawing release.
func (v DrawingVersion) Release() int {
	switch v {
	default:
		return 0
	case "MC0.0":
		return 1
	case "AC1.2":
		return 2
	case "AC1.4":
		return 3
	case "AC1.50":
		return 4
	case "AC2.10":
		return 5
	case "AC1002":
		return 6
	case "AC1003":
		return 7
	case "AC1004":
		return 8
	case "AC1006":
		return 9
	case "AC1009":
		return 10
	case "AC1012":
		return 11
	case "AC1014":
		return 12
	case "AC1015":
		return 13
	case "AC1018":
		return 14
	case "AC1021":
		return 15
	case "AC1024":
		return 16
	case "AC1027":
		return 17
	case "AC1032":
		return 18
	}
}

// ReleaseName returns a description of which versions of AutoCAD the drawing
// drawing is supported in.
func (v DrawingVersion) ReleaseName() string {
	switch v {
	case "MC0.0":
		return "Release 1.1"
	case "AC1.2":
		return "Release 1.2"
	case "AC1.4":
		return "Release 1.4"
	case "AC1.50":
		return "Release 2.0"
	case "AC2.10":
		return "Release 2.10"
	case "AC1002":
		return "Release 2.5"
	case "AC1003":
		return "Release 2.6"
	case "AC1004":
		return "Release 9"
	case "AC1006":
		return "Release 10"
	case "AC1009":
		return "Release 11/12 (LT R1/R2)"
	case "AC1012":
		return "Release 13 (LT95)"
	case "AC1014":
		return "Release 14, 14.01 (LT97/LT98)"
	case "AC1015":
		return "AutoCAD 2000/2000i/2002"
	case "AC1018":
		return "AutoCAD 2004/2005/2006"
	case "AC1021":
		return "AutoCAD 2007/2008/2009"
	case "AC1024":
		return "AutoCAD 2010/2011/2012"
	case "AC1027":
		return "AutoCAD 2013/2014/2015/2016/2017"
	case "AC1032":
		return "AutoCAD 2018/2019/2020"
	default:
		return ""
	}
}
