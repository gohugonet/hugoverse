package overlayfs

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"time"
)

// OverlayFs is a filesystem that overlays multiple filesystems.
// It's by default a read-only filesystem.
// For all operations, the filesystems are checked in order until found.
type OverlayFs struct {
	fss []afero.Fs
}

func (ofs *OverlayFs) Create(name string) (afero.File, error) {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Mkdir(name string, perm os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) MkdirAll(path string, perm os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Remove(name string) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) RemoveAll(path string) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Rename(oldname, newname string) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Name() string {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Chmod(name string, mode os.FileMode) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Chown(name string, uid, gid int) error {
	//TODO implement me
	panic("implement me")
}

func (ofs *OverlayFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	//TODO implement me
	panic("implement me")
}

// New creates a new OverlayFs with the given options.
func New(fss []afero.Fs) *OverlayFs {
	return &OverlayFs{
		fss: fss,
	}
}

func (ofs *OverlayFs) Append(fss ...afero.Fs) *OverlayFs {
	ofs.fss = append(ofs.fss, fss...)
	return ofs
}

// Open opens a file, returning it or an error, if any happens.
// If name is a directory, a *Dir is returned representing all directories matching name.
// Note that a *Dir must not be used after it's closed.
func (ofs *OverlayFs) Open(name string) (afero.File, error) {
	fmt.Println(">>> OverlayFs.Open", name)

	bfs, fi, err := ofs.stat(name)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		dir := getDir()
		dir.name = name

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

	f, err := bfs.Open(name)
	if err != nil {
		return nil, err
	}

	return &File{File: f}, err
}

func (ofs *OverlayFs) collectDirs(name string, withFs func(fs afero.Fs)) error {
	for _, fs := range ofs.fss {
		if err := ofs.collectDirsRecursive(fs, name, withFs); err != nil {
			return err
		}
	}
	return nil
}

func (ofs *OverlayFs) collectDirsRecursive(fs afero.Fs, name string, withFs func(fs afero.Fs)) error {
	if fi, err := fs.Stat(name); err == nil && fi.IsDir() {
		withFs(fs)
	}
	return nil
}

func (ofs *OverlayFs) Stat(name string) (os.FileInfo, error) {
	_, fi, err := ofs.stat(name)
	return fi, err
}

func (ofs *OverlayFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	_, fi, err := ofs.stat(name)
	return fi, false, err
}

func (ofs *OverlayFs) stat(name string) (afero.Fs, os.FileInfo, error) {
	for _, bfs := range ofs.fss {
		if fi, err := bfs.Stat(name); err == nil || !os.IsNotExist(err) {
			return bfs, fi, nil
		}
	}
	return nil, nil, os.ErrNotExist
}
