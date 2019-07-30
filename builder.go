// +build ignore

package main

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/josephspurrier/goversioninfo"
)

func main() {
	major, minor, patch, build := splitVersion(Version)
	fileVersion := goversioninfo.FileVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}
	vi := goversioninfo.VersionInfo{
		IconPath:     "icon.ico",
		ManifestPath: "cadscan.manifest",
		FixedFileInfo: goversioninfo.FixedFileInfo{
			FileVersion:    fileVersion,
			ProductVersion: fileVersion,
			FileFlagsMask:  "3f",
			FileFlags:      "00",
			FileOS:         "040004",
			FileType:       "01",
			FileSubType:    "00",
		},
		StringFileInfo: goversioninfo.StringFileInfo{
			CompanyName:      "SCJ Alliance",
			FileDescription:  "CAD File Scanner",
			FileVersion:      Version,
			OriginalFilename: "cadscan.exe",
			ProductName:      "CAD File Scanner",
			ProductVersion:   Version,
		},
		VarFileInfo: goversioninfo.VarFileInfo{
			Translation: goversioninfo.Translation{
				LangID:    goversioninfo.LngUSEnglish,
				CharsetID: goversioninfo.CsUnicode,
			},
		},
	}
	vi.Build()
	vi.Walk()
	vi.WriteSyso("cadscan.syso", runtime.GOARCH)
}

func splitVersion(version string) (major, minor, patch, build int) {
	parts := strings.Split(Version, ".")
	switch len(parts) {
	case 4:
		if val, err := strconv.Atoi(parts[3]); err == nil {
			build = val
		}
		fallthrough
	case 3:
		if val, err := strconv.Atoi(parts[2]); err == nil {
			patch = val
		}
		fallthrough
	case 2:
		if val, err := strconv.Atoi(parts[1]); err == nil {
			minor = val
		}
		fallthrough
	case 1:
		if val, err := strconv.Atoi(parts[0]); err == nil {
			major = val
		}
	}
	return
}
