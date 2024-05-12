package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"io/fs"
	"os"
)

type FileMetaInfo interface {
	fs.DirEntry
	os.FileInfo
	Meta() *FileMeta
}

func NewFileMetaInfo(fi FileNameIsDir, m *FileMeta) FileMetaInfo {
	if m == nil {
		panic("FileMeta must be set")
	}
	if fim, ok := fi.(MetaProvider); ok {
		m.Merge(fim.Meta())
	}
	switch v := fi.(type) {
	case fs.DirEntry:
		return &dirEntryMeta{DirEntry: v, m: m}
	case fs.FileInfo:
		return &dirEntryMeta{DirEntry: dirEntry{v}, m: m}
	case nil:
		return &dirEntryMeta{DirEntry: dirEntry{}, m: m}
	default:
		panic(fmt.Sprintf("Unsupported type: %T", fi))
	}
}

func DecorateFileInfo(fi FileNameIsDir, opener func() (afero.File, error), filename string, inMeta *FileMeta) FileMetaInfo {
	var meta *FileMeta
	var fim FileMetaInfo

	var ok bool
	if fim, ok = fi.(FileMetaInfo); ok {
		meta = fim.Meta()
	} else {
		meta = NewFileMeta()
		fim = NewFileMetaInfo(fi, meta)
	}

	if opener != nil {
		meta.OpenFunc = opener
	}

	nfilename := normalizeFilename(filename)
	if nfilename != "" {
		meta.Filename = nfilename
	}

	meta.Merge(inMeta)

	return fim
}
