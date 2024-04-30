package overlayfs

import (
	"fmt"
	"io"
	"io/fs"
	iofs "io/fs"
	"os"
	"sync"

	"github.com/spf13/afero"
)

var (
	_ FilesystemIterator = (*OverlayFs)(nil)
	_ afero.Fs           = (*OverlayFs)(nil)
	_ afero.Lstater      = (*OverlayFs)(nil)
	_ afero.File         = (*Dir)(nil)
	_ fs.ReadDirFile     = (*Dir)(nil)
)

// FilesystemIterator is an interface for iterating over the wrapped filesystems in order.
type FilesystemIterator interface {
	Filesystem(i int) afero.Fs
	NumFilesystems() int
}

// Options for the OverlayFs.
type Options struct {
	// The filesystems to overlay ordered in priority from left to right.
	Fss []afero.Fs

	// The OverlayFs is by default read-only, but you can nominate the first filesystem to be writable.
	FirstWritable bool

	// The DirsMerger is used to merge the contents of two directories.
	// If not provided, the defaultDirMerger is used.
	DirsMerger DirsMerger
}

// OverlayFs is a filesystem that overlays multiple filesystems.
// It's by default a read-only filesystem, but you can nominate the first filesystem to be writable.
// For all operations, the filesystems are checked in order until found.
// If a filesystem implementes FilesystemIterator, those filesystems will be checked before continuing.
type OverlayFs struct {
	fss []afero.Fs

	mergeDirs     DirsMerger
	firstWritable bool
}

// New creates a new OverlayFs with the given options.
func New(opts Options) *OverlayFs {
	if opts.DirsMerger == nil {
		opts.DirsMerger = defaultDirMerger
	}

	return &OverlayFs{
		fss:           opts.Fss,
		mergeDirs:     opts.DirsMerger,
		firstWritable: opts.FirstWritable,
	}
}

// Append creates a shallow copy of the filesystem and appends the given filesystems to it.
func (ofs OverlayFs) Append(fss ...afero.Fs) *OverlayFs {
	ofs.fss = append(ofs.fss, fss...)
	return &ofs
}

// WithDirsMerger creates a shallow copy of the filesystem and sets the DirsMerger.
func (ofs OverlayFs) WithDirsMerger(d DirsMerger) *OverlayFs {
	ofs.mergeDirs = d
	return &ofs
}

// Filesystem returns filesystem with index i, nil if not found.
func (ofs *OverlayFs) Filesystem(i int) afero.Fs {
	if i >= len(ofs.fss) {
		return nil
	}
	return ofs.fss[i]
}

// NumFilesystems returns the number of filesystems in this composite filesystem.
func (ofs *OverlayFs) NumFilesystems() int {
	return len(ofs.fss)
}

// Name returns the name of this filesystem.
func (ofs *OverlayFs) Name() string {
	return "overlayfs"
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
	if fsi, ok := fs.(FilesystemIterator); ok {
		for i := 0; i < fsi.NumFilesystems(); i++ {
			if err := ofs.collectDirsRecursive(fsi.Filesystem(i), name, withFs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ofs *OverlayFs) stat(name string, lstatIfPossible bool) (afero.Fs, os.FileInfo, bool, error) {
	for _, fs := range ofs.fss {
		if fs2, fi, ok, err := ofs.statRecursive(fs, name, lstatIfPossible); err == nil || !os.IsNotExist(err) {
			return fs2, fi, ok, err
		}
	}
	return nil, nil, false, os.ErrNotExist
}

func (ofs *OverlayFs) statRecursive(fs afero.Fs, name string, lstatIfPossible bool) (afero.Fs, os.FileInfo, bool, error) {
	if lstatIfPossible {
		if lfs, ok := fs.(afero.Lstater); ok {
			fi, ok, err := lfs.LstatIfPossible(name)
			if err == nil || !os.IsNotExist(err) {
				return fs, fi, ok, err
			}
		} else if fi, err := fs.Stat(name); err == nil || !os.IsNotExist(err) {
			return fs, fi, false, err
		}
	} else if fi, err := fs.Stat(name); err == nil || !os.IsNotExist(err) {
		return fs, fi, false, err
	}
	if fsi, ok := fs.(FilesystemIterator); ok {
		for i := 0; i < fsi.NumFilesystems(); i++ {
			if fs2, fi, ok, err := ofs.statRecursive(fsi.Filesystem(i), name, lstatIfPossible); err == nil || !os.IsNotExist(err) {
				return fs2, fi, ok, err
			}
		}
	}
	return nil, nil, false, os.ErrNotExist
}

func (ofs *OverlayFs) writeFs() afero.Fs {
	if len(ofs.fss) == 0 {
		panic("overlayfs: there are no filesystems to write to")
	}
	return ofs.fss[0]
}

// DirsMerger is used to merge two directories.
type DirsMerger func(lofi, bofi []fs.DirEntry) []fs.DirEntry

var defaultDirMerger = func(lofi, bofi []fs.DirEntry) []fs.DirEntry {
	for _, bofi := range bofi {
		var found bool
		for _, lofi := range lofi {
			if bofi.Name() == lofi.Name() {
				found = true
				break
			}
		}
		if !found {
			lofi = append(lofi, bofi)
		}
	}
	return lofi
}

var dirPool = &sync.Pool{
	New: func() any {
		return &Dir{}
	},
}

func getDir() *Dir {
	return dirPool.Get().(*Dir)
}

func releaseDir(dir *Dir) {
	dir.fss = dir.fss[:0]
	dir.fis = dir.fis[:0]
	dir.dirOpeners = dir.dirOpeners[:0]
	dir.info = nil
	dir.offset = 0
	dir.name = ""
	dir.err = nil
	dirPool.Put(dir)
}

// OpenDir opens a new Dir with dirs to be merged by the given merge func.
// If merge is nil, a default DirsMerger is used.
func OpenDir(
	merge DirsMerger,
	// Used to stat the directory.
	info func() (os.FileInfo, error),
	// Used to open the directories to be merged.
	dirOpeners ...func() (afero.File, error),
) (*Dir, error) {
	if merge == nil {
		merge = defaultDirMerger
	}
	if info == nil {
		panic("overlayfs: info must not be nil")
	}
	if len(dirOpeners) == 0 {
		panic("overlayfs: dirOpeners must not be empty")
	}

	dir := getDir()
	dir.dirOpeners = dirOpeners
	dir.info = info
	dir.merge = merge
	return dir, nil
}

// Dir is an afero.File that represents list of directories that will be merged in Readdir and Readdirnames.
type Dir struct {
	// It's either a named directory in a slice of filesystems or a slice of directories.
	name string
	fss  []afero.Fs

	// Set if fss is not set.
	dirOpeners []func() (afero.File, error)
	info       func() (os.FileInfo, error)

	merge DirsMerger

	err    error
	offset int
	fis    []fs.DirEntry
}

// Readdir implements afero.File.Readdir.
// If n > 0, Readdir returns at most n.
// Note that Dir also implements fs.ReadDirFile, which is more efficient.
func (d *Dir) Readdir(n int) ([]os.FileInfo, error) {
	dirEntries, err := d.ReadDir(n)
	if err != nil {
		return nil, err
	}
	fis := make([]os.FileInfo, len(dirEntries))
	for i, dirEntry := range dirEntries {
		fi, err := dirEntry.Info()
		if err != nil {
			return nil, err
		}
		fis[i] = fi
	}
	return fis, nil
}

// ReadDir implements fs.ReadDirFile.
func (d *Dir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.err != nil {
		return nil, d.err
	}
	if d.isClosed() {
		return nil, os.ErrClosed
	}

	if d.offset == 0 {
		readDir := func(fs afero.Fs, f afero.File) error {
			var err error
			if f == nil {
				f, err = fs.Open(d.name)
				if err != nil {
					return err
				}
			}
			defer f.Close()

			var dirEntries []iofs.DirEntry

			if rdf, ok := f.(iofs.ReadDirFile); ok {
				dirEntries, err = rdf.ReadDir(-1)
				if err != nil {
					return err
				}
			} else {
				var fis []os.FileInfo
				fis, err = f.Readdir(-1)
				if err != nil {
					return err
				}
				dirEntries = make([]iofs.DirEntry, len(fis))
				for i, fi := range fis {
					dirEntries[i] = dirEntry{fi}
				}
			}

			d.fis = d.merge(d.fis, dirEntries)
			return nil
		}

		for _, fs := range d.fss {
			if err := readDir(fs, nil); err != nil {
				return nil, err
			}
		}
		for _, open := range d.dirOpeners {
			f, err := open()
			if err != nil {
				return nil, err
			}
			if err := readDir(nil, f); err != nil {
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
		fisc := make([]fs.DirEntry, len(fis))
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

	fisc := make([]fs.DirEntry, len(fis[:n]))
	copy(fisc, fis[:n])

	return fisc, nil
}

// Readdirnames implements afero.File.Readdirnames.
// If n > 0, Readdirnames returns at most n.
func (d *Dir) Readdirnames(n int) ([]string, error) {
	if d.isClosed() {
		return nil, os.ErrClosed
	}

	fis, err := d.ReadDir(n)
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
	if d.isClosed() {
		return nil, os.ErrClosed
	}
	if d.info != nil {
		return d.info()
	}
	return d.fss[0].Stat(d.name)
}

// Close implements afero.File.Close.
// Note that d must not be used after it is closed,
// as the object may be reused.
func (d *Dir) Close() error {
	releaseDir(d)
	return nil
}

// Name implements afero.File.Name.
func (d *Dir) Name() string {
	return d.name
}

func (d *Dir) notSupported() string {
	return fmt.Sprintf("operation not supported on directory %q", d.name)
}

// Read is not supported.
func (d *Dir) Read(p []byte) (n int, err error) {
	panic(d.notSupported())
}

// ReadAt is not supported.
func (d *Dir) ReadAt(p []byte, off int64) (n int, err error) {
	panic(d.notSupported())
}

// Seek is not supported.
func (d *Dir) Seek(offset int64, whence int) (int64, error) {
	panic(d.notSupported())
}

// Write is not supported.
func (d *Dir) Write(p []byte) (n int, err error) {
	panic(d.notSupported())
}

// WriteAt is not supported.
func (d *Dir) WriteAt(p []byte, off int64) (n int, err error) {
	panic(d.notSupported())
}

// Sync is not supported.
func (d *Dir) Sync() error {
	panic(d.notSupported())
}

// Truncate is not supported.
func (d *Dir) Truncate(size int64) error {
	panic(d.notSupported())
}

// WriteString is not supported.
func (d *Dir) WriteString(s string) (ret int, err error) {
	panic(d.notSupported())
}

func (d *Dir) isClosed() bool {
	return len(d.fss) == 0 && len(d.dirOpeners) == 0
}

// dirEntry is an adapter from os.FileInfo to fs.DirEntry
type dirEntry struct {
	fs.FileInfo
}

var _ fs.DirEntry = dirEntry{}

func (d dirEntry) Type() fs.FileMode { return d.FileInfo.Mode().Type() }

func (d dirEntry) Info() (fs.FileInfo, error) { return d.FileInfo, nil }
