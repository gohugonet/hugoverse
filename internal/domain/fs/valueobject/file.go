package valueobject

import "github.com/spf13/afero"

func NewFile(file afero.File) *File {
	return &File{file}
}

type File struct {
	afero.File
}
