package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"io/fs"
)

type DirOpener func() ([]fs.DirEntry, error)

type DirFile struct {
	*File

	virtualOpener DirOpener

	filter func([]fs.DirEntry) ([]fs.DirEntry, error)
}

func NewDirFileWithVirtualOpener(f *File, opener DirOpener) *DirFile {
	return &DirFile{File: f, virtualOpener: opener}
}

func NewDirFile(file afero.File, meta FileMeta, fs afero.Fs) *DirFile {
	return &DirFile{File: &File{File: file, FileMeta: meta, fs: fs, isDir: true}, virtualOpener: nil}
}

func (f *DirFile) ReadDir(count int) ([]fs.DirEntry, error) {
	if f.File.File != nil {
		fis, err := f.File.ReadDir(count)
		fmt.Println("ReadDir 666: ", fis, err)
		if err != nil {
			return nil, err
		}

		if f.filter != nil {
			fis, err = f.filter(fis)
			if err != nil {
				return nil, err
			}
		}

		return fis, nil
	}

	return f.readVirtualDir()
}

func (f *DirFile) readVirtualDir() ([]fs.DirEntry, error) {
	if f.virtualOpener != nil {
		return f.virtualOpener()
	}
	return nil, fmt.Errorf("virtual dir opener not found")
}
