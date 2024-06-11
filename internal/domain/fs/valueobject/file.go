package valueobject

import (
	"github.com/spf13/afero"
)

func NewFile(file afero.File, filename string) *File {
	return &File{File: file, filename: filename}
}

type File struct {
	afero.File
	filename string
}

func (f *File) Close() error {
	if f.File == nil {
		return nil
	}
	return f.File.Close()
}
