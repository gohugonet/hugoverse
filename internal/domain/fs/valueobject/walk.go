package valueobject

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	iofs "io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"

	"github.com/gohugonet/hugoverse/internal/domain/fs"
)

type Walkway struct {
	// The filesystem to walk.
	Fs afero.Fs

	// The root to start from in Fs.
	Root string

	cb *WalkCallback

	logger loggers.Logger

	// Prevent a walkway to be walked more than once.
	walked bool

	// Config from client.
	cfg WalkwayConfig
}

type WalkCallback struct {
	// Will be called in order.
	HookPre  fs.WalkHook // Optional.
	WalkFn   fs.WalkFunc
	HookPost fs.WalkHook // Optional.
}

func (w *WalkCallback) WalkHook() fs.WalkFunc { return w.WalkFn }
func (w *WalkCallback) PreHook() fs.WalkHook  { return w.HookPre }
func (w *WalkCallback) PostHook() fs.WalkHook { return w.HookPost }

type WalkwayConfig struct {
	// One or both of these may be pre-set.
	Info       fs.FileMetaInfo            // The start info.
	DirEntries []fs.FileMetaInfo          // The start info's dir entries.
	IgnoreFile func(filename string) bool // Optional

	// Some optional flags.
	FailOnNotExist bool // If set, return an error if a directory is not found.
	SortDirEntries bool // If set, sort the dir entries by Name before calling the WalkFn, default is ReaDir order.
}

func NewWalkway(fs afero.Fs, cb fs.WalkCallback) (*Walkway, error) {
	if fs == nil {
		return nil, errors.New("fs must be set")
	}
	if cb.WalkHook() == nil {
		return nil, errors.New("walkFn must be set")
	}

	return &Walkway{
		Fs: fs,
		cb: &WalkCallback{
			HookPre:  cb.PreHook(),
			WalkFn:   cb.WalkHook(),
			HookPost: cb.PostHook(),
		},

		logger: loggers.NewDefault(),
	}, nil
}

func (w *Walkway) WalkWith(root string, cfg WalkwayConfig) error {
	w.cfg = cfg
	w.Root = root
	return w.Walk()
}

func (w *Walkway) Walk() error {
	if w.walked {
		panic("this walkway is already walked")
	}
	w.walked = true

	if w.Fs == NoOpFs {
		return nil
	}

	return w.walk(w.Root, w.cfg.Info, w.cfg.DirEntries)
}

// checkErr returns true if the error is handled.
func (w *Walkway) checkErr(filename string, err error) bool {
	if herrors.IsNotExist(err) && !w.cfg.FailOnNotExist {
		// The file may be removed in process.
		// This may be a ERROR situation, but it is not possible
		// to determine as a general case.
		w.logger.Warnf("File %q not found, skipping.", filename)
		return true
	}

	return false
}

// walk recursively descends path, calling walkFn.
func (w *Walkway) walk(path string, info fs.FileMetaInfo, dirEntries []fs.FileMetaInfo) error {
	pathRel := strings.TrimPrefix(path, w.Root)

	if info == nil {
		var err error
		fi, err := w.Fs.Stat(path)
		if err != nil {
			if path == w.Root && herrors.IsNotExist(err) {
				return nil
			}
			if w.checkErr(path, err) {
				return nil
			}
			return fmt.Errorf("walk: stat: %s", err)
		}
		info = fi.(fs.FileMetaInfo)
	}

	err := w.cb.WalkFn(path, info)
	if err != nil {
		if info.IsDir() && errors.Is(err, filepath.SkipDir) {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	if dirEntries == nil {
		f, err := w.Fs.Open(path)
		fmt.Println("555 open file: ", path)
		if err != nil {
			if w.checkErr(path, err) {
				return nil
			}
			return fmt.Errorf("walk: open: path: %q filename: %q: %s", path, info.FileName(), err)
		}
		fis, err := f.(iofs.ReadDirFile).ReadDir(-1)

		_ = f.Close()
		if err != nil {
			if w.checkErr(path, err) {
				return nil
			}
			return fmt.Errorf("walk: Readdir: %w", err)
		}

		dirEntries = DirEntriesToFileMetaInfos(fis)
		for _, fi := range dirEntries {
			if fi.Path() == nil {
				fi.SetPath(paths.Parse("", filepath.Join(pathRel, fi.Name())))
			}
		}

		if w.cfg.SortDirEntries {
			sort.Slice(dirEntries, func(i, j int) bool {
				return dirEntries[i].Name() < dirEntries[j].Name()
			})
		}

	}

	if w.cfg.IgnoreFile != nil {
		n := 0
		for _, fi := range dirEntries {
			if !w.cfg.IgnoreFile(fi.FileName()) {
				dirEntries[n] = fi
				n++
			}
		}
		dirEntries = dirEntries[:n]
	}

	if w.cb.HookPre != nil {
		var err error
		dirEntries, err = w.cb.HookPre(info, path, dirEntries)
		if err != nil {
			if errors.Is(err, filepath.SkipDir) {
				return nil
			}
			return err
		}
	}

	for _, fim := range dirEntries {
		nextPath := filepath.Join(path, fim.Name())
		err := w.walk(nextPath, fim, nil)
		if err != nil {
			if !fim.IsDir() || !errors.Is(err, filepath.SkipDir) {
				return err
			}
		}
	}

	if w.cb.HookPost != nil {
		var err error
		dirEntries, err = w.cb.HookPost(info, path, dirEntries)
		if err != nil {
			if errors.Is(err, filepath.SkipDir) {
				return nil
			}
			return err
		}
	}
	return nil
}

func DirEntriesToFileMetaInfos(fis []iofs.DirEntry) []fs.FileMetaInfo {
	fims := make([]fs.FileMetaInfo, len(fis))
	for i, v := range fis {
		fim := v.(fs.FileMetaInfo)
		fims[i] = fim
	}
	return fims
}

// WalkFn is the walk func for WalkFilesystems.
type WalkFn func(fs afero.Fs) bool

// WalkFilesystems walks fs recursively and calls fn.
// If fn returns true, walking is stopped.
func WalkFilesystems(afs afero.Fs, fn WalkFn) bool {
	if fn(afs) {
		return true
	}

	if afs, ok := afs.(fs.FilesystemUnwrapper); ok {
		if WalkFilesystems(afs.UnwrapFilesystem(), fn) {
			return true
		}
	} else if bfs, ok := afs.(fs.FilesystemsUnwrapper); ok {
		for _, sf := range bfs.UnwrapFilesystems() {
			if WalkFilesystems(sf, fn) {
				return true
			}
		}
	} else if cfs, ok := afs.(overlayfs.FilesystemIterator); ok {
		for i := 0; i < cfs.NumFilesystems(); i++ {
			if WalkFilesystems(cfs.Filesystem(i), fn) {
				return true
			}
		}
	}

	return false
}
