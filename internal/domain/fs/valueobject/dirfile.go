package valueobject

import (
	"github.com/spf13/afero"
	"io/fs"
)

type DirOpener func(name string) ([]fs.DirEntry, error)

type DirFile struct {
	File

	opener DirOpener
}

func NewDirFile(file afero.File) *DirFile {
	return &DirFile{File: File{File: file}, opener: nil}
}

func NewDirFileWithOpener(file afero.File, opener DirOpener) *DirFile {
	f := NewDirFile(file)
	f.opener = opener
	return f
}

func (f *DirFile) ReadDir(count int) ([]fs.DirEntry, error) {
	if f.File.File != nil {
		fis, err := f.File.File.(fs.ReadDirFile).ReadDir(count)
		if err != nil {
			return nil, err
		}

		var result []fs.DirEntry
		for _, fi := range fis {
			fim := NewDirInfo(fi)
			result = append(result, fim)
		}
		return result, nil
	}

	return f.opener(f.Name())
}
