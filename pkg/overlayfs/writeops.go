package overlayfs

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

// Chmod changes the mode of the named file to mode.
func (ofs *OverlayFs) Chmod(name string, mode os.FileMode) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Chmod(name, mode)
}

// Chown changes the uid and gid of the named file.
func (ofs *OverlayFs) Chown(name string, uid, gid int) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Chown(name, uid, gid)
}

// Chtimes changes the access and modification times of the named file
func (ofs *OverlayFs) Chtimes(name string, atime, mtime time.Time) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Chtimes(name, atime, mtime)
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (ofs *OverlayFs) Mkdir(name string, perm os.FileMode) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Mkdir(name, perm)
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (ofs *OverlayFs) MkdirAll(path string, perm os.FileMode) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().MkdirAll(path, perm)
}

// OpenFile opens a file using the given flags and the given mode.
func (ofs *OverlayFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		if !ofs.firstWritable {
			return nil, os.ErrPermission
		}
		return ofs.writeFs().OpenFile(name, flag, perm)
	}
	return ofs.Open(name)
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (ofs *OverlayFs) Remove(name string) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Remove(name)
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (ofs *OverlayFs) RemoveAll(path string) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().RemoveAll(path)
}

// Rename renames a file.
func (ofs *OverlayFs) Rename(oldname, newname string) error {
	if !ofs.firstWritable {
		return os.ErrPermission
	}
	return ofs.writeFs().Rename(oldname, newname)
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (ofs *OverlayFs) Create(name string) (afero.File, error) {
	if !ofs.firstWritable {
		return nil, os.ErrPermission
	}
	return ofs.writeFs().Create(name)
}
