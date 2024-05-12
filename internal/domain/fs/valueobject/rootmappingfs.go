package valueobject

import (
	"fmt"
	dfs "github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/htime"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/radixtree"
	"github.com/spf13/afero"
	"golang.org/x/text/unicode/norm"
	iofs "io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// A RootMappingFs maps several roots into one.
// Note that the root of this filesystem
// is directories only, and they will be returned
// in Readdir and Readdirnames
// in the order given.
type RootMappingFs struct {
	afero.Fs
	RootMapToReal *radixtree.Tree
	RealMapToRoot *radixtree.Tree
}

func (m *RootMappingFs) Dirs(base string) ([]FileMetaInfo, error) {
	base = dfs.FilepathSeparator + base
	roots := m.getRootsWithPrefix(base)

	if roots == nil {
		return nil, nil
	}

	fss := make([]FileMetaInfo, len(roots))
	for i, r := range roots {
		bfs := afero.NewBasePathFs(m.Fs, r.To)
		bfs = decorateDirs(bfs, r.Meta)

		fi, err := bfs.Stat("")
		if err != nil {
			return nil, fmt.Errorf("RootMappingFs.Dirs: %w", err)
		}

		fss[i] = fi.(FileMetaInfo)
	}

	return fss, nil
}

func (m *RootMappingFs) getRootsWithPrefix(prefix string) []RootMapping {
	return GetRms(m.RootMapToReal, prefix) // /content
}

// Open opens the named file for reading.
func (m *RootMappingFs) Open(name string) (afero.File, error) {
	fis, err := m.doLstat(name)
	if err != nil {
		return nil, err
	}

	return m.newUnionFile(fis...)
}

func (m *RootMappingFs) newUnionFile(fis ...FileMetaInfo) (afero.File, error) {
	if len(fis) == 1 {
		return fis[0].Meta().Open()
	}

	if !fis[0].IsDir() {
		// Pick the last file mount.
		return fis[len(fis)-1].Meta().Open()
	}

	openers := make([]func() (afero.File, error), len(fis))
	for i := len(fis) - 1; i >= 0; i-- {
		fi := fis[i]
		openers[i] = func() (afero.File, error) {
			meta := fi.Meta()
			f, err := meta.Open()
			if err != nil {
				return nil, err
			}
			return &rootMappingDir{DirOnlyOps: f, fs: m, name: meta.Name, meta: meta}, nil
		}
	}

	merge := func(lofi, bofi []iofs.DirEntry) []iofs.DirEntry {
		// Ignore duplicate directory entries
		for _, fi1 := range bofi {
			var found bool
			for _, fi2 := range lofi {
				if !fi2.IsDir() {
					continue
				}
				if fi1.Name() == fi2.Name() {
					found = true
					break
				}
			}
			if !found {
				lofi = append(lofi, fi1)
			}
		}

		return lofi
	}

	info := func() (os.FileInfo, error) {
		return fis[0], nil
	}

	return overlayfs.OpenDir(merge, info, openers...)
}

// Stat returns the os.FileInfo structure describing a given file.  If there is
// an error, it will be of type *os.PathError.
func (m *RootMappingFs) Stat(name string) (os.FileInfo, error) {
	fi, _, err := m.LstatIfPossible(name)
	return fi, err
}

// LstatIfPossible returns the os.FileInfo structure describing a given file.
func (m *RootMappingFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fis, err := m.doLstat(name)
	if err != nil {
		return nil, false, err
	}
	return fis[0], false, nil
}

func (m *RootMappingFs) doLstat(name string) ([]FileMetaInfo, error) {
	name = m.cleanName(name)
	key := paths.FilePathSeparator + name

	roots := m.getRoot(key)

	if roots == nil {
		// Find any real files or directories with this key.
		_, roots := m.getRoots(key)
		if roots == nil {
			return nil, &os.PathError{Op: "LStat", Path: name, Err: os.ErrNotExist}
		}

		var err error
		var fis []FileMetaInfo

		for _, rm := range roots {
			var fi FileMetaInfo
			fi, _, err = m.statRoot(rm, name)
			if err == nil {
				fis = append(fis, fi)
			}
		}

		if fis != nil {
			return fis, nil
		}

		if err == nil {
			err = &os.PathError{Op: "LStat", Path: name, Err: err}
		}

		return nil, err
	}

	return []FileMetaInfo{roots[0].Fi}, nil
}

func (m *RootMappingFs) cleanName(name string) string {
	return strings.Trim(filepath.Clean(name), paths.FilePathSeparator)
}

func (m *RootMappingFs) getRoot(key string) []RootMapping {
	v, found := m.RootMapToReal.Get(key)
	if !found {
		return nil
	}

	return v.([]RootMapping)
}

func (m *RootMappingFs) getRoots(key string) (string, []RootMapping) {
	s, v, found := m.RootMapToReal.LongestPrefix(key)
	if !found || (s == paths.FilePathSeparator && key != paths.FilePathSeparator) {
		return "", nil
	}
	return s, v.([]RootMapping)
}

func (m *RootMappingFs) statRoot(root RootMapping, name string) (FileMetaInfo, bool, error) {
	filename := root.Filename(name)

	fi, b, err := LstatIfPossible(m.Fs, filename) // source fs
	if err != nil {
		return nil, b, err
	}

	var opener func() (afero.File, error)
	if fi.IsDir() {
		// Make sure metadata gets applied in Readdir.
		opener = m.realDirOpener(filename, root.Meta)
	} else {
		// Opens the real file directly.
		opener = func() (afero.File, error) {
			return m.Fs.Open(filename)
		}
	}

	return DecorateFileInfo(fi, opener, "", root.Meta), b, nil
}

func (m *RootMappingFs) realDirOpener(name string, meta *FileMeta) func() (afero.File, error) {
	return func() (afero.File, error) {
		f, err := m.Fs.Open(name)
		if err != nil {
			return nil, err
		}
		return &rootMappingDir{name: name, meta: meta, fs: m, DirOnlyOps: f}, nil
	}
}

func (m *RootMappingFs) collectDirEntries(prefix string) ([]iofs.DirEntry, error) {
	prefix = paths.FilePathSeparator + m.cleanName(prefix)

	var fis []iofs.DirEntry

	seen := make(map[string]bool) // Prevent duplicate directories
	level := strings.Count(prefix, paths.FilePathSeparator)

	collectDir := func(rm RootMapping, fi FileMetaInfo) error {
		f, err := fi.Meta().Open()
		if err != nil {
			return err
		}
		direntries, err := f.(iofs.ReadDirFile).ReadDir(-1)
		if err != nil {
			f.Close()
			return err
		}

		for _, fi := range direntries {
			meta := fi.(FileMetaInfo).Meta()
			meta.Merge(rm.Meta)

			if fi.IsDir() {
				name := fi.Name()
				if seen[name] {
					continue
				}
				seen[name] = true
				opener := func() (afero.File, error) {
					return m.Open(filepath.Join(rm.From, name))
				}
				fi = newDirNameOnlyFileInfo(name, meta, opener)
			}

			fis = append(fis, fi)
		}

		f.Close()

		return nil
	}

	// First add any real files/directories.
	rms := m.getRoot(prefix)
	for _, rm := range rms {
		if err := collectDir(rm, rm.Fi); err != nil {
			return nil, err
		}
	}

	// Next add any file mounts inside the given directory.
	prefixInside := prefix + paths.FilePathSeparator
	m.RootMapToReal.WalkPrefix(prefixInside, func(s string, v any) bool {
		if (strings.Count(s, paths.FilePathSeparator) - level) != 1 {
			// This directory is not part of the current, but we
			// need to include the first name part to make it
			// navigable.
			path := strings.TrimPrefix(s, prefixInside)
			parts := strings.Split(path, paths.FilePathSeparator)
			name := parts[0]

			if seen[name] {
				return false
			}
			seen[name] = true
			opener := func() (afero.File, error) {
				return m.Open(path)
			}

			fi := newDirNameOnlyFileInfo(name, nil, opener)
			fis = append(fis, fi)

			return false
		}

		rms := v.([]RootMapping)
		for _, rm := range rms {
			if !rm.Fi.IsDir() {
				// A single file mount
				fis = append(fis, rm.Fi)
				continue
			}
			name := filepath.Base(rm.From)
			if seen[name] {
				continue
			}
			seen[name] = true

			opener := func() (afero.File, error) {
				return m.Open(rm.From)
			}

			fi := newDirNameOnlyFileInfo(name, rm.Meta, opener)

			fis = append(fis, fi)

		}

		return false
	})

	// Finally add any ancestor dirs with files in this directory.
	//ancestors := m.getAncestors(prefix)
	//for _, root := range ancestors {
	//	subdir := strings.TrimPrefix(prefix, root.key)
	//	for _, rm := range root.roots {
	//		if rm.fi.IsDir() {
	//			fi, err := rm.fi.Meta().JoinStat(subdir)
	//			if err == nil {
	//				if err := collectDir(rm, fi); err != nil {
	//					return nil, err
	//				}
	//			}
	//		}
	//	}
	//}

	return fis, nil
}

func newDirNameOnlyFileInfo(name string, meta *FileMeta, fileOpener func() (afero.File, error)) FileMetaInfo {
	name = normalizeFilename(name)
	_, base := filepath.Split(name)

	m := meta.Copy()
	if m.Filename == "" {
		m.Filename = name
	}
	m.OpenFunc = fileOpener
	m.IsOrdered = false

	return NewFileMetaInfo(
		NewDirNameOnlyFI(base, htime.Now()),
		m,
	)
}

func normalizeFilename(filename string) string {
	if filename == "" {
		return ""
	}
	if runtime.GOOS == "darwin" {
		// When a file system is HFS+, its filepath is in NFD form.
		return norm.NFC.String(filename)
	}
	return filename
}

func (m *RootMappingFs) Mounts(base string) ([]FileMetaInfo, error) {
	base = dfs.FilepathSeparator + m.cleanName(base)
	roots := m.getRootsWithPrefix(base)

	if roots == nil {
		return nil, nil
	}

	fss := make([]FileMetaInfo, len(roots))
	for i, r := range roots {
		// TODO: ignore single file mount
		bfs := NewBasePathFs(m.Fs, r.To)
		fs := bfs
		// TODO: ignore InclusionFilter
		fs = decorateDirs(fs, r.Meta)
		fi, err := fs.Stat("")
		if err != nil {
			return nil, fmt.Errorf("RootMappingFs.Dirs: %w", err)
		}
		fss[i] = fi.(FileMetaInfo)
	}

	return fss, nil
}

// func (fs *RootMappingFs) ReverseStat(filename string) ([]FileMetaInfo, error)
func (m *RootMappingFs) ReverseLookup(filename string) ([]ComponentPath, error) {
	return m.ReverseLookupComponent("", filename)
}

func (m *RootMappingFs) ReverseLookupComponent(component, filename string) ([]ComponentPath, error) {
	filename = m.cleanName(filename)
	key := dfs.FilepathSeparator + filename

	s, roots := m.getRootsReverse(key)

	if len(roots) == 0 {
		return nil, nil
	}

	var cps []ComponentPath

	base := strings.TrimPrefix(key, s)
	dir, name := filepath.Split(base)

	for _, first := range roots {
		if component != "" && first.FromBase != component {
			continue
		}

		var filename string
		if first.Meta.Rename != nil { // TODO, Rename is nil
			// Single file mount.
			if newname, ok := first.Meta.Rename(name, true); ok {
				filename = dfs.FilepathSeparator + filepath.Join(first.path, dir, newname)
			} else {
				continue
			}
		} else {
			// Now we know that this file _could_ be in this fs.
			filename = dfs.FilepathSeparator + filepath.Join(first.path, dir, name)
		}

		cps = append(cps, ComponentPath{
			Component: first.FromBase,
			Path:      paths.ToSlashTrimLeading(filename),
			Lang:      first.Meta.Lang,
		})
	}

	return cps, nil
}

func (m *RootMappingFs) getRootsReverse(key string) (string, []RootMapping) {
	tree := m.RealMapToRoot
	s, v, found := tree.LongestPrefix(key)
	if !found {
		return "", nil
	}
	return s, v.([]RootMapping)
}
