package valueobject

import "os"

type fileInfoMeta struct {
	os.FileInfo

	m *FileMeta
}

func (fi *fileInfoMeta) Meta() *FileMeta {
	return fi.m
}
