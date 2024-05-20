package valueobject

import (
	"github.com/spf13/afero"
	iofs "io/fs"
	"strings"
)

func NewFile(file afero.File) *File {
	return &File{file}
}

type File struct {
	afero.File
}

func (f *File) Close() error {
	if f.File == nil {
		return nil
	}
	return f.File.Close()
}

func (f *File) Name() string {
	return f.File.Name()
}

func (f *File) ReadDir(count int) ([]iofs.DirEntry, error) {
	if f.DirOnlyOps != nil {
		fis, err := f.DirOnlyOps.(iofs.ReadDirFile).ReadDir(count)
		if err != nil {
			return nil, err
		}

		var result []iofs.DirEntry
		for _, fi := range fis {
			fim := decorateFileInfo(fi, nil, "", f.meta)
			meta := fim.Meta()
			if f.meta.InclusionFilter.Match(strings.TrimPrefix(meta.Filename, meta.SourceRoot), fim.IsDir()) {
				result = append(result, fim)
			}
		}
		return result, nil
	}

	return f.fs.collectDirEntries(f.name)
}
