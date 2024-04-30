package overlayfs

import (
	"os"

	"github.com/spf13/afero"
)

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (ofs *OverlayFs) Stat(name string) (os.FileInfo, error) {
	_, fi, _, err := ofs.stat(name, false)
	return fi, err
}

// LstatIfPossible will call Lstat if the filesystem iself is, or it delegates to, the os filesystem.
// Else it will call Stat.
func (ofs *OverlayFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	_, fi, ok, err := ofs.stat(name, false)
	return fi, ok, err
}

// Open opens a file, returning it or an error, if any happens.
// If name is a directory, a *Dir is returned representing all directories matching name.
// Note that a *Dir must not be used after it's closed.
func (ofs *OverlayFs) Open(name string) (afero.File, error) {
	fs, fi, _, err := ofs.stat(name, false)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		dir := getDir()
		dir.name = name
		dir.merge = ofs.mergeDirs
		if err := ofs.collectDirs(name, func(fs afero.Fs) {
			dir.fss = append(dir.fss, fs)
		}); err != nil {
			dir.Close()
			return nil, err
		}

		if len(dir.fss) == 0 {
			// They mave been deleted.
			dir.Close()
			return nil, os.ErrNotExist
		}

		if len(dir.fss) == 1 {
			// Optimize for the common case.
			d, err := dir.fss[0].Open(name)
			dir.Close()
			return d, err
		}

		return dir, nil
	}

	return fs.Open(name)
}
