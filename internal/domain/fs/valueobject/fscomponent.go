package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
)

type ComponentFs struct {
	afero.Fs

	// The component name, e.g. "content", "layouts" etc.
	Component string

	OverlayFs afero.Fs
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
		filename: name,
		PathInfo: cfs.pathResolver(name),
	}

	if fi.IsDir() {
		meta.OpenFunc = func() (afero.File, error) {
			return cfs.Open(name)
		}
	}

	return NewFileInfoWithMeta(fi, meta), nil
}

func (cfs *ComponentFs) pathResolver(name string) *paths.Path {
	p := paths.NewPathParser()

	return p.Parse(cfs.Component, name)
}

func (cfs *ComponentFs) Open(name string) (afero.File, error) {
	f, err := cfs.Fs.Open(name)
	if err != nil {
		return nil, err
	}

	fi, err := cfs.Fs.Stat(name)
	if err != nil {
		f.Close()
		return nil, err
	}

	if fi.IsDir() {
		df := NewDirFile(f, name, cfs)
		df.filter = symlinkFilter
		df.sorter = sorter
		df.pathResolver = cfs.pathResolver

		return df, nil
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

func sorter(fis []iofs.DirEntry) []iofs.DirEntry {
	sort.Slice(fis, func(i, j int) bool {
		fimi, fimj := fis[i].(fs.FileMetaInfo), fis[j].(fs.FileMetaInfo)
		if fimi.IsDir() != fimj.IsDir() {
			return fimi.IsDir()
		}

		return fimi.Name() < fimj.Name()
	})

	return fis
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
