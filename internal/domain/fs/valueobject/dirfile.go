package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"io/fs"
	"path/filepath"
)

type DirOpener func() ([]fs.DirEntry, error)

type DirFile struct {
	*File
	fs afero.Fs

	virtualOpener DirOpener

	filter func([]fs.DirEntry) ([]fs.DirEntry, error)
	sorter func([]fs.DirEntry) []fs.DirEntry
}

func NewDirFileWithVirtualOpener(f *File, opener DirOpener) *DirFile {
	return &DirFile{File: f, virtualOpener: opener}
}

func NewDirFile(file afero.File, meta FileMeta, fs afero.Fs) *DirFile {
	return &DirFile{File: &File{File: file, FileMeta: meta}, fs: fs, virtualOpener: nil}
}

func (f *DirFile) ReadDir(count int) ([]fs.DirEntry, error) {
	if f.File.File != nil {
		fis, err := f.File.File.(fs.ReadDirFile).ReadDir(count)
		if err != nil {
			return nil, err
		}

		if f.filter != nil {
			fis, err = f.filter(fis)
			if err != nil {
				return nil, err
			}
		}

		var result []fs.DirEntry
		for _, fi := range fis {
			filename := filepath.Join(f.File.filename, fi.Name())

			meta := &FileMeta{
				filename: filename,
				OpenFunc: func() (afero.File, error) {
					return f.fs.Open(filename)
				},
			}
			meta.Merge(&f.File.FileMeta)

			fim, err := NewFileInfoWithDirEntryMeta(fi, meta)

			if err != nil {
				return nil, err
			}
			result = append(result, fim)
		}

		if f.sorter != nil {
			result = f.sorter(result)
		}

		return result, nil
	}

	return f.readVirtualDir()
}

func (f *DirFile) readVirtualDir() ([]fs.DirEntry, error) {
	if f.virtualOpener != nil {
		return f.virtualOpener()
	}
	return nil, fmt.Errorf("virtual dir opener not found")
}
