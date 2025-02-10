package valueobject

import (
	"errors"
	"fmt"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/overlayfs"
	"github.com/spf13/afero"
	iofs "io/fs"
	"path/filepath"
	"sort"

	"github.com/mdfriday/hugoverse/internal/domain/fs"
)

type Walkway struct {
	// The filesystem to walk.
	Fs afero.Fs

	// The ComponentRoot to start from in Fs.
	Root string

	cb fs.WalkCallback

	logger loggers.Logger

	// Prevent a walkway to be walked more than once.
	walked bool

	// Config from client.
	cfg fs.WalkwayConfig
}

func NewWalkway(fs afero.Fs, cb fs.WalkCallback) (*Walkway, error) {
	if fs == nil {
		return nil, errors.New("fs must be set")
	}
	if cb.WalkFn == nil {
		return nil, errors.New("walkFn must be set")
	}

	return &Walkway{
		Fs: fs,
		cb: cb,

		logger: loggers.NewDefault(),
	}, nil
}

func (w *Walkway) WalkWith(root string, cfg fs.WalkwayConfig) error {
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

	if dfs, ok := afs.(fs.FilesystemUnwrapper); ok {
		if WalkFilesystems(dfs.UnwrapFilesystem(), fn) {
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
