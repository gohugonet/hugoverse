package valueobject

import "io/fs"

type DirInfo struct {
	fs.DirEntry
	fi *FileInfo
}

func NewDirInfo(d fs.DirEntry) *DirInfo {
	return &DirInfo{DirEntry: d, fi: &FileInfo{FileMeta: FileMeta{}}}
}

func (fi *DirInfo) Meta() *FileMeta {
	return fi.fi.Meta()
}
