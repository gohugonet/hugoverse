package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/spf13/afero"
	iofs "io/fs"
	"os"
	"path/filepath"
)

type ComponentFs struct {
	afero.Fs

	// The component name, e.g. "content", "layouts" etc.
	Component string

	OverlayFs afero.Fs

	log loggers.Logger
}

func NewComponentFs(component string, overlayFs *overlayfs.OverlayFs) *ComponentFs {
	return &ComponentFs{
		Component: component,
		OverlayFs: overlayFs,
		Fs:        NewBasePathFs(overlayFs, component),

		log: loggers.NewDefault(),
	}
}

func (cfs *ComponentFs) UnwrapFilesystem() afero.Fs {
	return cfs.Fs
}

func (cfs *ComponentFs) Stat(name string) (os.FileInfo, error) {
	fi, err := cfs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}

	meta := &FileMeta{
		filename:  name,
		component: cfs.Component,
	}

	if fi.IsDir() {
		meta.OpenFunc = func() (afero.File, error) {
			return cfs.Open(name)
		}
	}

	return NewFileInfoWithMeta(fi, meta), nil
}

func (cfs *ComponentFs) Open(name string) (afero.File, error) {
	cfs.log.Printf("Open (ComponentFs): %s", name)

	f, err := cfs.Fs.Open(name)
	if err != nil {
		return nil, err
	}

	if baseFile, ok := f.(*afero.BasePathFile); ok {
		if dirFile, ok := baseFile.File.(*DirFile); ok {
			dirFile.filter = symlinkFilter
			return dirFile, nil
		}
	}

	return f, nil
}

func symlinkFilter(fis []iofs.DirEntry) ([]iofs.DirEntry, error) {
	var filtered []iofs.DirEntry
	for _, fi := range fis {
		// IsDir will always be false for symlinks.
		keep := fi.IsDir()
		if !keep {
			// This is unfortunate, but is the only way to determine if it is a symlink.
			info, err := fi.Info()
			if err != nil {
				if herrors.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			if !isSymlink(info) {
				keep = true
			}
		}
		if keep {
			filtered = append(filtered, fi)
		}
	}

	return filtered, nil
}

func isSymlink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink != 0
}

// RealDirs gets a list of absolute paths to directories starting from the given
// path.
func (cfs *ComponentFs) RealDirs(from string) []string {
	var dirnames []string
	for _, m := range cfs.mounts() {
		if !m.IsDir() {
			continue
		}
		dirname := filepath.Join(m.FileName(), from)
		dirnames = append(dirnames, dirname)
	}
	return dirnames
}

func (cfs *ComponentFs) mounts() []fs.FileMetaInfo {
	var m []fs.FileMetaInfo
	WalkFilesystems(cfs.Fs, func(fs afero.Fs) bool {
		if rfs, ok := fs.(*RootMappingFs); ok {
			mounts, err := rfs.Mounts(cfs.Component)
			if err == nil {
				m = append(m, mounts...)
			}
		}
		return false
	})

	return m
}

// MakePathRelative creates a relative path from the given filename.
func (cfs *ComponentFs) MakePathRelative(filename string, checkExists bool) (string, bool) {
	cps, err := cfs.ReverseLookup(filename, checkExists)
	if err != nil {
		panic(err)
	}
	if len(cps) == 0 {
		return "", false
	}

	return filepath.FromSlash(cps[0].Path), true
}

// ReverseLookup returns the component paths for the given filename.
func (cfs *ComponentFs) ReverseLookup(filename string, checkExists bool) ([]ComponentPath, error) {
	var cps []ComponentPath
	WalkFilesystems(cfs.Fs, func(fs afero.Fs) bool {
		if rfs, ok := fs.(ReverseLookupProvider); ok {
			if c, err := rfs.ReverseLookupComponent(cfs.Component, filename); err == nil {
				if checkExists {
					n := 0
					for _, cp := range c {
						if _, err := cfs.Fs.Stat(filepath.FromSlash(cp.Path)); err == nil {
							c[n] = cp
							n++
						}
					}
					c = c[:n]
				}
				cps = append(cps, c...)
			}
		}
		return false
	})
	return cps, nil
}
