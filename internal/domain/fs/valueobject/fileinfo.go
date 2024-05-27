package valueobject

import (
	"io/fs"
	"os"
)

func NewFileInfo(fi os.FileInfo, filename string) *FileInfo {
	info := &FileInfo{
		FileInfo: fi,
		FileMeta: &FileMeta{
			filename: filename,
			OpenFunc: nil,
		},
	}

	if fim, ok := fi.(MetaProvider); ok {
		info.FileMeta.Merge(fim.Meta())
	}

	return info
}

func NewFileInfoWithDirEntry(de fs.DirEntry) (*FileInfo, error) {
	fi, err := de.Info()
	if err != nil {
		return nil, err
	}
	return NewFileInfo(fi, fi.Name()), nil
}

func NewFileInfoWithDirEntryOpener(de fs.DirEntry, opener FileOpener) (*FileInfo, error) {
	fi, err := de.Info()
	if err != nil {
		return nil, err
	}
	return NewFileInfoWithOpener(fi, fi.Name(), opener), nil
}

func NewFileInfoWithDirEntryMeta(de fs.DirEntry, meta *FileMeta) (*FileInfo, error) {
	fi, err := de.Info()
	if err != nil {
		return nil, err
	}
	return NewFileInfoWithMeta(fi, meta), nil
}

func NewFileInfoWithOpener(fi os.FileInfo, filename string, opener FileOpener) *FileInfo {
	info := NewFileInfo(fi, filename)
	info.OpenFunc = opener

	return info
}

func NewFileInfoWithMeta(fi os.FileInfo, meta *FileMeta) *FileInfo {
	info := NewFileInfo(fi, meta.FileName())
	info.FileMeta = meta

	return info
}

type FileInfo struct {
	fs.FileInfo
	*FileMeta
}

func (fi *FileInfo) Meta() *FileMeta {
	return fi.FileMeta
}

func (fi *FileInfo) Type() fs.FileMode {
	return fi.FileInfo.Mode()
}

func (fi *FileInfo) Info() (fs.FileInfo, error) {
	return fi.FileInfo, nil
}
