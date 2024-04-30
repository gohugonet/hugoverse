package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"io/fs"
)

// LanguageDirsMerger implements the overlayfs.DirsMerger func, which is used
// to merge two directories.
var LanguageDirsMerger overlayfs.DirsMerger = func(lofi, bofi []fs.DirEntry) []fs.DirEntry {
	for _, fi1 := range bofi {
		fim1 := fi1.(FileMetaInfo)
		var found bool
		for _, fi2 := range lofi {
			fim2 := fi2.(FileMetaInfo)
			if fi1.Name() == fi2.Name() && fim1.Meta().Lang == fim2.Meta().Lang {
				found = true
				break
			}
		}
		if !found {
			lofi = append(lofi, fi1)
		}
	}

	return lofi
}

// AppendDirsMerger merges two directories keeping all regular files
// with the first slice as the base.
// Duplicate directories in the second slice will be ignored.
// This strategy is used for the i18n and data fs where we need all entries.
var AppendDirsMerger overlayfs.DirsMerger = func(lofi, bofi []fs.DirEntry) []fs.DirEntry {
	for _, fi1 := range bofi {
		var found bool
		// Remove duplicate directories.
		if fi1.IsDir() {
			for _, fi2 := range lofi {
				if fi2.IsDir() && fi2.Name() == fi1.Name() {
					found = true
					break
				}
			}
		}
		if !found {
			lofi = append(lofi, fi1)
		}
	}

	return lofi
}
