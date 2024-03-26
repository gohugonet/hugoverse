package valueobject

import (
	"os"
	"time"
)

func NewDirNameOnlyFI(name string, modTime time.Time) *DirNameOnlyFileInfo {
	return &DirNameOnlyFileInfo{name: name, modTime: modTime}
}

type DirNameOnlyFileInfo struct {
	name    string
	modTime time.Time
}

func (fi *DirNameOnlyFileInfo) Name() string {
	return fi.name
}

func (fi *DirNameOnlyFileInfo) Size() int64 {
	panic("not implemented")
}

func (fi *DirNameOnlyFileInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (fi *DirNameOnlyFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *DirNameOnlyFileInfo) IsDir() bool {
	return true
}

func (fi *DirNameOnlyFileInfo) Sys() any {
	return nil
}
