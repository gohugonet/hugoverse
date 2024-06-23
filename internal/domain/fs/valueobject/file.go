package valueobject

import (
	"github.com/spf13/afero"
)

type File struct {
	afero.File
	FileMeta
}

func (f *File) Close() error {
	if f.File == nil {
		return nil
	}
	return f.File.Close()
}

func NewFile(file afero.File, filename string) *File {
	return &File{File: file, FileMeta: FileMeta{filename: filename}}
}
