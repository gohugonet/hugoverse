package overlayfs

import (
	"io/fs"
	"os"
)

type File struct {
	fs.File
}

func (f File) ReadAt(p []byte, off int64) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (f File) Write(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (f File) WriteAt(p []byte, off int64) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (f File) Name() string {
	//TODO implement me
	panic("implement me")
}

func (f File) Readdir(count int) ([]os.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f File) Readdirnames(n int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f File) Sync() error {
	//TODO implement me
	panic("implement me")
}

func (f File) Truncate(size int64) error {
	//TODO implement me
	panic("implement me")
}

func (f File) WriteString(s string) (ret int, err error) {
	//TODO implement me
	panic("implement me")
}
