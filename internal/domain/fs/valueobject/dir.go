package valueobject

import (
	"io"
	"io/fs"
	"os"
	"sync"
	"time"
)

// DirOnlyOps is a subset of the afero.File interface covering
// the methods needed for directory operations.
type DirOnlyOps interface {
	io.Closer
	Name() string
	Readdir(count int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
	Stat() (os.FileInfo, error)
}

type FileNameIsDir interface {
	Name() string
	IsDir() bool
}

type dirEntryMeta struct {
	fs.DirEntry
	m *FileMeta

	fi     fs.FileInfo
	fiInit sync.Once
}

func (fi *dirEntryMeta) Meta() *FileMeta {
	return fi.m
}

// Filename returns the full filename.
func (fi *dirEntryMeta) Filename() string {
	return fi.m.Filename
}

func (fi *dirEntryMeta) fileInfo() fs.FileInfo {
	var err error
	fi.fiInit.Do(func() {
		fi.fi, err = fi.DirEntry.Info()
	})
	if err != nil {
		panic(err)
	}
	return fi.fi
}

func (fi *dirEntryMeta) Size() int64 {
	return fi.fileInfo().Size()
}

func (fi *dirEntryMeta) Mode() fs.FileMode {
	return fi.fileInfo().Mode()
}

func (fi *dirEntryMeta) ModTime() time.Time {
	return fi.fileInfo().ModTime()
}

func (fi *dirEntryMeta) Sys() any {
	return fi.fileInfo().Sys()
}

// Name returns the file's name.
func (fi *dirEntryMeta) Name() string {
	if name := fi.m.Name; name != "" {
		return name
	}
	return fi.DirEntry.Name()
}

// dirEntry is an adapter from os.FileInfo to fs.DirEntry
type dirEntry struct {
	fs.FileInfo
}

var _ fs.DirEntry = dirEntry{}

func (d dirEntry) Type() fs.FileMode { return d.FileInfo.Mode().Type() }

func (d dirEntry) Info() (fs.FileInfo, error) { return d.FileInfo, nil }

func DirEntriesToFileMetaInfos(fis []fs.DirEntry) []FileMetaInfo {
	fims := make([]FileMetaInfo, len(fis))
	for i, v := range fis {
		fim := v.(FileMetaInfo)
		fims[i] = fim
	}
	return fims
}
