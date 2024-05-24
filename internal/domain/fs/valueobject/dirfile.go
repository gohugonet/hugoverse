package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"io/fs"
)

type DirOpener func(name string) ([]fs.DirEntry, error)

type DirFile struct {
	*File

	virtualOpener DirOpener
}

func NewDirFileWithFile(f *File, opener DirOpener) *DirFile {
	return &DirFile{File: f, virtualOpener: opener}
}

func NewDirFile(file afero.File) *DirFile {
	return &DirFile{File: &File{File: file}, virtualOpener: nil}
}

func (f *DirFile) ReadDir(count int) ([]fs.DirEntry, error) {
	if f.File.File != nil {
		fis, err := f.File.File.(fs.ReadDirFile).ReadDir(count)
		if err != nil {
			return nil, err
		}

		var result []fs.DirEntry
		for _, fi := range fis {
			fim, err := NewFileInfoWithDirEntry(fi)
			if err != nil {
				return nil, err
			}
			result = append(result, fim)
		}
		return result, nil
	}

	return f.readVirtualDir()
}

func (f *DirFile) readVirtualDir() ([]fs.DirEntry, error) {
	if f.virtualOpener != nil {
		return f.virtualOpener(f.filename)
	}
	return nil, fmt.Errorf("virtual dir opener not found")
}
