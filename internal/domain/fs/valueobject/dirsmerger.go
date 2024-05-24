package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	iofs "io/fs"
)

// LanguageDirsMerger implements the overlayfs.DirsMerger func, which is used
// to merge two directories.
var LanguageDirsMerger overlayfs.DirsMerger = func(lofi, bofi []iofs.DirEntry) []iofs.DirEntry {
	for _, fi1 := range bofi {
		var found bool
		for _, fi2 := range lofi {
			if fi1.Name() == fi2.Name() { // ignore lang
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
var AppendDirsMerger overlayfs.DirsMerger = func(lofi, bofi []iofs.DirEntry) []iofs.DirEntry {
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
