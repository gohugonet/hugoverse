package overlayfs

import (
	"github.com/spf13/afero"
	"io"
	"os"
	"sync"
)

// DirsMerger is used to merge two directories.
type DirsMerger func(lofi, bofi []os.FileInfo) []os.FileInfo

// Dir is an afero.File that represents list of directories that will be merged in Readdir and Readdirnames.
type Dir struct {
	name  string
	fss   []afero.Fs
	merge DirsMerger

	err    error
	offset int
	fis    []os.FileInfo
}

var dirPool = &sync.Pool{
	New: func() any {
		return &Dir{}
	},
}

func getDir() *Dir {
	return dirPool.Get().(*Dir)
}

// Close implements afero.File.Close.
// Note that d must not be used after it is closed,
// as the object may be reused.
func (d *Dir) Close() error {
	releaseDir(d)
	return nil
}

func releaseDir(dir *Dir) {
	dir.fss = dir.fss[:0]
	dir.fis = dir.fis[:0]
	dir.offset = 0
	dir.name = ""
	dir.err = nil
	dirPool.Put(dir)
}

// Readdir implements afero.File.Readdir.
// If n > 0, Readdir returns at most n.
func (d *Dir) Readdir(n int) ([]os.FileInfo, error) {
	if d.err != nil {
		return nil, d.err
	}
	if len(d.fss) == 0 {
		return nil, os.ErrClosed
	}

	if d.offset == 0 {
		readDir := func(fs afero.Fs) error {
			f, err := fs.Open(d.name)
			if err != nil {
				return err
			}
			defer f.Close()
			fi, err := f.Readdir(-1)
			if err != nil {
				return err
			}
			d.fis = d.merge(d.fis, fi)
			return nil
		}

		for _, fs := range d.fss {
			if err := readDir(fs); err != nil {
				return nil, err
			}
		}
	}

	fis := d.fis[d.offset:]

	if n <= 0 {
		d.err = io.EOF
		if d.offset > 0 && len(fis) == 0 {
			return nil, d.err
		}
		fisc := make([]os.FileInfo, len(fis))
		copy(fisc, fis)
		return fisc, nil
	}

	if len(fis) == 0 {
		d.err = io.EOF
		return nil, d.err
	}

	if n > len(d.fis) {
		n = len(d.fis)
	}

	defer func() { d.offset += n }()

	fisc := make([]os.FileInfo, len(fis[:n]))
	copy(fisc, fis[:n])

	return fisc, nil

}

// Readdirnames implements afero.File.Readdirnames.
// If n > 0, Readdirnames returns at most n.
func (d *Dir) Readdirnames(n int) ([]string, error) {
	if len(d.fss) == 0 {
		return nil, os.ErrClosed
	}

	fis, err := d.Readdir(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(fis))
	for i, fi := range fis {
		names[i] = fi.Name()
	}
	return names, nil
}

// Stat implements afero.File.Stat.
func (d *Dir) Stat() (os.FileInfo, error) {
	if len(d.fss) == 0 {
		return nil, os.ErrClosed
	}
	return d.fss[0].Stat(d.name)
}

// Name implements afero.File.name.
func (d *Dir) Name() string {
	return d.name
}

// Read is not supported.
func (d *Dir) Read(p []byte) (n int, err error) {
	panic("not supported")
}

// ReadAt is not supported.
func (d *Dir) ReadAt(p []byte, off int64) (n int, err error) {
	panic("not supported")
}

// Seek is not supported.
func (d *Dir) Seek(offset int64, whence int) (int64, error) {
	panic("not supported")
}

// Write is not supported.
func (d *Dir) Write(p []byte) (n int, err error) {
	panic("not supported")
}

// WriteAt is not supported.
func (d *Dir) WriteAt(p []byte, off int64) (n int, err error) {
	panic("not supported")
}

// Sync is not supported.
func (d *Dir) Sync() error {
	panic("not supported")
}

// Truncate is not supported.
func (d *Dir) Truncate(size int64) error {
	panic("not supported")
}

// WriteString is not supported.
func (d *Dir) WriteString(s string) (ret int, err error) {
	panic("not supported")
}
