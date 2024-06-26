package valueobject

import (
	"errors"
	"github.com/spf13/afero"
	"os"
	"time"
)

var (
	errNoOp          = errors.New("this operation is not supported")
	_       afero.Fs = (*noOpFs)(nil)

	// NoOpFs provides a no-op filesystem that implements the afero.Fs
	// interface.
	NoOpFs = &noOpFs{}
)

type noOpFs struct{}

func (fs noOpFs) Create(name string) (afero.File, error) {
	panic(errNoOp)
}

func (fs noOpFs) Mkdir(name string, perm os.FileMode) error {
	return nil
}

func (fs noOpFs) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (fs noOpFs) Open(name string) (afero.File, error) {
	return nil, os.ErrNotExist
}

func (fs noOpFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, os.ErrNotExist
}

func (fs noOpFs) Remove(name string) error {
	return nil
}

func (fs noOpFs) RemoveAll(path string) error {
	return nil
}

func (fs noOpFs) Rename(oldname string, newname string) error {
	panic(errNoOp)
}

func (fs noOpFs) Stat(name string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func (fs noOpFs) Name() string {
	return "noOpFs"
}

func (fs noOpFs) Chmod(name string, mode os.FileMode) error {
	panic(errNoOp)
}

func (fs noOpFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	panic(errNoOp)
}

func (fs *noOpFs) Chown(name string, uid int, gid int) error {
	panic(errNoOp)
}

// noOpRegularFileOps implements the non-directory operations of a afero.File
// panicking for all operations.
type noOpRegularFileOps struct{}

func (f *noOpRegularFileOps) Read(p []byte) (n int, err error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) ReadAt(p []byte, off int64) (n int, err error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Seek(offset int64, whence int) (int64, error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Write(p []byte) (n int, err error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) WriteAt(p []byte, off int64) (n int, err error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Readdir(count int) ([]os.FileInfo, error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Readdirnames(n int) ([]string, error) {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Sync() error {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) Truncate(size int64) error {
	panic(errNoOp)
}

func (f *noOpRegularFileOps) WriteString(s string) (ret int, err error) {
	panic(errNoOp)
}
