package valueobject

import (
	"errors"
	iofs "io/fs"
	"os"
)

type rootMappingDir struct {
	*noOpRegularFileOps
	DirOnlyOps
	fs   *RootMappingFs
	name string
	meta *FileMeta
}

func (f *rootMappingDir) Close() error {
	if f.DirOnlyOps == nil {
		return nil
	}
	return f.DirOnlyOps.Close()
}

func (f *rootMappingDir) Name() string {
	return f.name
}

func (f *rootMappingDir) ReadDir(count int) ([]iofs.DirEntry, error) {
	if f.DirOnlyOps != nil {
		fis, err := f.DirOnlyOps.(iofs.ReadDirFile).ReadDir(count)
		if err != nil {
			return nil, err
		}

		var result []iofs.DirEntry
		for _, fi := range fis {
			fim := DecorateFileInfo(fi, nil, "", f.meta)
			// TODO: ignore InclusionFilter
			result = append(result, fim)
		}
		return result, nil
	}

	return f.fs.collectDirEntries(f.name)
}

// Sentinel error to signal that a file is a directory.
var errIsDir = errors.New("isDir")

func (f *rootMappingDir) Stat() (iofs.FileInfo, error) {
	return nil, errIsDir
}

func (f *rootMappingDir) Readdir(count int) ([]os.FileInfo, error) {
	dirEntry, err := f.ReadDir(count)
	if err != nil {
		return nil, err
	}
	var result []os.FileInfo
	for _, d := range dirEntry {
		result = append(result, d.(os.FileInfo))
	}
	return result, nil
}

// Note that Readdirnames preserves the order of the underlying filesystem(s),
// which is usually directory order.
func (f *rootMappingDir) Readdirnames(count int) ([]string, error) {
	dirs, err := f.ReadDir(count)
	if err != nil {
		return nil, err
	}
	return dirEntriesToNames(dirs), nil
}

func dirEntriesToNames(fis []iofs.DirEntry) []string {
	names := make([]string, len(fis))
	for i, d := range fis {
		names[i] = d.Name()
	}
	return names
}
