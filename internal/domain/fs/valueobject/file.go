package valueobject

import (
	"github.com/spf13/afero"
	"io/fs"
	"path/filepath"
)

type File struct {
	afero.File
	FileMeta

	fs    afero.Fs
	isDir bool
}

func (f *File) isNop() bool {
	return f.File == nil
}

func (f *File) Close() error {
	if f.File == nil {
		return nil
	}
	return f.File.Close()
}

func (f *File) ReadDir(count int) ([]fs.DirEntry, error) {
	var result []fs.DirEntry

	if f.isDir {
		fis, err := f.File.Readdir(count)
		if err != nil {
			return result, err
		}

		for _, fi := range fis {
			filename := filepath.Join(f.filename, fi.Name())

			meta := &FileMeta{
				filename: filename,
				OpenFunc: func() (afero.File, error) {
					return f.fs.Open(filename)
				},
			}
			meta.Merge(&f.FileMeta)

			fim := NewFileInfoWithNewMeta(fi, meta)

			result = append(result, fim)
		}
	}

	return result, nil
}

func NewFile(file afero.File, filename string) *File {
	return &File{File: file, FileMeta: FileMeta{filename: filename}}
}
