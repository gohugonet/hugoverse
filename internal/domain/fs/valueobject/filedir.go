package valueobject

import "github.com/spf13/afero"

type DirFile struct {
	File
}

func NewDirFile(file afero.File) *DirFile {
	return &DirFile{File: File{File: file}}
}
