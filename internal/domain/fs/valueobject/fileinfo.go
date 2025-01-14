package valueobject

import (
	"io/fs"
	"os"
)

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

func NewFileInfoWithMeta(fi os.FileInfo, meta *FileMeta) *FileInfo {
	info := &FileInfo{
		FileInfo: fi,
		FileMeta: meta,
	}

	if fim, ok := fi.(MetaProvider); ok {
		info.FileMeta.Merge(fim.Meta())
	}

	return info
}

func NewFileInfoWithName(filename string) *FileInfo {
	vf, _ := GetVirtualFileInfo()

	return &FileInfo{
		FileInfo: vf,
		FileMeta: &FileMeta{
			filename: filename,
			OpenFunc: nil,
		},
	}
}

func NewFileInfoWithContent(content string) *FileInfo {
	vf, _ := GetVirtualFileInfoWithContent(content)
	info, _ := vf.GetFileInfo()

	return &FileInfo{
		FileInfo: info,
		FileMeta: &FileMeta{
			filename:      vf.FullName(),
			OpenFunc:      vf.Open,
			ComponentRoot: "content",
			ComponentDir:  "content",
		},
	}
}

func NewFileInfoWithOpener(fi os.FileInfo, filename string, opener FileOpener) *FileInfo {
	info := NewFileInfo(fi, filename)
	info.OpenFunc = opener

	return info
}

func NewFileInfoWithRoot(fi os.FileInfo, filename, root string, opener FileOpener) *FileInfo {
	info := NewFileInfoWithOpener(fi, filename, opener)
	info.FileMeta.ComponentRoot = root

	return info
}

func NewFileInfoWithNewMeta(fi os.FileInfo, meta *FileMeta) *FileInfo {
	info := NewFileInfo(fi, meta.FileName())
	info.FileMeta = meta

	return info
}
