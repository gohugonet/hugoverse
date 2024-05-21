package valueobject

import (
	"errors"
	"github.com/spf13/afero"
	"io/fs"
	"os"
)

func NewFileInfo(fi os.FileInfo, filename string) *FileInfo {
	info := &FileInfo{
		FileInfo: fi,
		FileMeta: FileMeta{
			Name:     filename,
			OpenFunc: nil,
		},
	}

	if fim, ok := fi.(MetaProvider); ok {
		info.FileMeta.Merge(fim.Meta())
	}

	return info
}

func NewFileInfoWithOpener(fi os.FileInfo, filename string, opener FileOpener) *FileInfo {
	info := NewFileInfo(fi, filename)
	info.OpenFunc = opener

	return info
}

type FileInfo struct {
	fs.FileInfo
	FileMeta
}

func (fi *FileInfo) Name() string {
	return fi.FileInfo.Name()
}

func (fi *FileInfo) Open() (afero.File, error) {
	if fi.OpenFunc == nil {
		return nil, errors.New("OpenFunc not set")
	}
	return fi.OpenFunc()
}

func (fi *FileInfo) Meta() *FileMeta {
	return &fi.FileMeta
}
