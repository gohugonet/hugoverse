package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	"github.com/gohugonet/hugoverse/pkg/radixtree"
	"github.com/spf13/afero"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
)

func NewRootMappingFs(fs afero.Fs, rms ...RootMapping) (*RootMappingFs, error) {
	rootMapToReal := radixtree.New()
	realMapToRoot := radixtree.New()

	addMapping := func(key string, rm RootMapping, to *radixtree.Tree) {
		var mappings []RootMapping
		v, found := to.Get(key)
		if found {
			// There may be more than one language pointing to the same root.
			mappings = v.([]RootMapping)
		}
		mappings = append(mappings, rm)
		to.Insert(key, mappings)
	}

	for _, rm := range rms {
		rm.clean()

		rm.FromBase = files.ResolveComponentFolder(rm.From)
		if rm.FromBase == "" {
			panic(" rm.FromBase is empty")
		}

		fi, err := fs.Stat(rm.To)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		// Do not support single file mount

		if !fi.IsDir() {
			panic("single file mount not supported yet, " + rm.To)
		}
		// fi: baseFs.Stat
		rm.ToFi = NewFileInfo(fi, rm.To)

		virKey := mapKey(rm.From)
		addMapping(virKey, rm, rootMapToReal)

		relKey := rm.To
		if !strings.HasPrefix(relKey, paths.FilePathSeparator) {
			relKey = mapKey(relKey)
		}

		addMapping(relKey, rm, realMapToRoot)
	}

	rfs := &RootMappingFs{
		Fs:            fs,
		rootMapToReal: rootMapToReal,
		realMapToRoot: realMapToRoot,
	}

	return rfs, nil
}

type RootMappingFs struct {
	afero.Fs
	rootMapToReal *radixtree.Tree
	realMapToRoot *radixtree.Tree
}

func (rmfs *RootMappingFs) UnwrapFilesystem() afero.Fs {
	return rmfs.Fs
}

// Stat returns the os.FileInfo structure describing a given file.  If there is
// an error, it will be of type *os.PathError.
func (rmfs *RootMappingFs) Stat(name string) (os.FileInfo, error) {
	fis, err := rmfs.doStat(name)
	if err != nil {
		return nil, err
	}
	// first win
	return fis[0], nil
}

func (rmfs *RootMappingFs) doStat(name string) ([]fs.FileMetaInfo, error) {
	fis, err := rmfs.doDoStat(name)
	if err != nil {
		return nil, err
	}
	// Sanity check. Check that all is either file or directories.
	var isDir, isFile bool
	for _, fi := range fis {
		if fi.IsDir() {
			isDir = true
		} else {
			isFile = true
		}
	}
	if isDir && isFile {
		// For now.
		return nil, os.ErrNotExist
	}

	return fis, nil
}

func (rmfs *RootMappingFs) doDoStat(name string) ([]fs.FileMetaInfo, error) {
	name = cleanName(name)
	key := mapKey(name)

	roots := rmfs.getRoot(key)

	if roots == nil {
		if rmfs.hasPrefix(key) {
			// We have directories mounted below this.
			// Make it look like a directory.
			panic("single file mount not supported yet")
		}

		// Find any real directories with this key.
		_, roots := rmfs.getRoots(key)
		if roots == nil {
			return nil, &os.PathError{Op: "LStat", Path: name, Err: os.ErrNotExist}
		}

		var err error
		var fis []fs.FileMetaInfo

		for _, rm := range roots {
			var fi fs.FileMetaInfo
			fi, err = rmfs.statRoot(rm, name)
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

	return []fs.FileMetaInfo{NewFileInfoWithOpener(roots[0].ToFi, name,
		func() (afero.File, error) {
			return NewDirFileWithVirtualOpener(
				&File{File: nil, filename: name},
				func() ([]iofs.DirEntry, error) {
					return rmfs.collectRootDirEntries(name)
				}), nil
		})}, nil
}

func (rmfs *RootMappingFs) statRoot(root RootMapping, filename string) (fs.FileMetaInfo, error) {
	filename = root.absFilename(filename)
	fi, err := rmfs.Fs.Stat(filename)
	if err != nil {
		return nil, err
	}

	var opener func() (afero.File, error)
	if !fi.IsDir() {
		opener = func() (afero.File, error) {
			return rmfs.Fs.Open(filename)
		}
	} else {
		// Make sure metadata gets applied in ReadDir.
		opener = func() (afero.File, error) {
			f, err := rmfs.Fs.Open(filename)
			if err != nil {
				return nil, err
			}

			df := NewDirFile(f, filename, rmfs)
			return df, nil
		}
	}

	fim := NewFileInfoWithOpener(fi, filename, opener)

	return fim, nil
}

func (rmfs *RootMappingFs) getRoot(key string) []RootMapping {
	v, found := rmfs.rootMapToReal.Get(key)
	if !found {
		return nil
	}

	return v.([]RootMapping)
}

func (rmfs *RootMappingFs) hasPrefix(prefix string) bool {
	hasPrefix := false
	rmfs.rootMapToReal.WalkPrefix(prefix, func(b string, v any) bool {
		hasPrefix = true
		return true
	})

	return hasPrefix
}

func (rmfs *RootMappingFs) getRoots(key string) (string, []RootMapping) {
	tree := rmfs.rootMapToReal
	levels := strings.Count(key, paths.FilePathSeparator)
	seen := make(map[RootMapping]bool)

	var roots []RootMapping
	var s string

	for {
		var found bool
		ss, vv, found := tree.LongestPrefix(key)

		if !found || (levels < 2 && ss == key) {
			break
		}

		for _, rm := range vv.([]RootMapping) {
			if !seen[rm] {
				seen[rm] = true
				roots = append(roots, rm)
			}
		}
		s = ss

		// We may have more than one root for this key, so walk up.
		oldKey := key
		key = filepath.Dir(key)
		if key == oldKey {
			break
		}
	}

	return s, roots
}

func (rmfs *RootMappingFs) collectRootDirEntries(prefix string) ([]iofs.DirEntry, error) {
	prefix = mapKey(prefix)

	var fis []iofs.DirEntry

	seen := make(map[string]bool) // Prevent duplicate directories
	level := strings.Count(prefix, paths.FilePathSeparator)

	collectDir := func(rm RootMapping, fi *FileInfo) error {
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
			if fi.IsDir() {
				name := fi.Name()
				if seen[name] {
					continue
				}
				seen[name] = true
				opener := func() (afero.File, error) {
					return rmfs.Open(rmfs.virtualPath(rm.From, name))
				}
				fi, err = NewFileInfoWithDirEntryOpener(fi, opener)
				if err != nil {
					return err
				}
			}
			fis = append(fis, fi)
		}

		f.Close()

		return nil
	}

	// First add any real files/directories.
	rms := rmfs.getRoot(prefix)
	for _, rm := range rms {
		if err := collectDir(rm, rm.ToFi); err != nil {
			return nil, err
		}
	}

	// Next add any file mounts inside the given directory.
	prefixInside := prefix + paths.FilePathSeparator
	rmfs.rootMapToReal.WalkPrefix(prefixInside, func(s string, v any) bool {
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
				return rmfs.Open(path)
			}

			fi := NewFileInfoWithOpener(nil, name, opener)
			fis = append(fis, fi)

			return false
		}

		rms := v.([]RootMapping)
		for _, rm := range rms {
			name := filepath.Base(rm.From)
			if seen[name] {
				continue
			}
			seen[name] = true
			opener := func() (afero.File, error) {
				return rmfs.Open(rm.From)
			}
			fi := NewFileInfoWithOpener(nil, name, opener)
			fis = append(fis, fi)
		}

		return false
	})

	// Finally add any ancestor dirs with files in this directory.
	ancestors := rmfs.getAncestors(prefix)
	for _, root := range ancestors {
		fmt.Println("warning: refer to ancestors ", root.key, root.roots)
	}

	return fis, nil
}

type keyRootMappings struct {
	key   string
	roots []RootMapping
}

func (rmfs *RootMappingFs) getAncestors(prefix string) []keyRootMappings {
	var roots []keyRootMappings
	rmfs.rootMapToReal.WalkPath(prefix, func(s string, v any) bool {
		if strings.HasPrefix(prefix, s+paths.FilePathSeparator) {
			roots = append(roots, keyRootMappings{
				key:   s,
				roots: v.([]RootMapping),
			})
		}
		return false
	})

	return roots
}

func (rmfs *RootMappingFs) virtualPath(rmFrom, name string) string {
	return filepath.Join(rmFrom, name)
}

// Open opens the named file for reading.
func (rmfs *RootMappingFs) Open(name string) (afero.File, error) {
	fis, err := rmfs.doStat(name)
	if err != nil {
		return nil, err
	}

	return rmfs.newUnionFile(fis...)
}

func (rmfs *RootMappingFs) newUnionFile(fis ...fs.FileMetaInfo) (afero.File, error) {
	if len(fis) == 1 {
		return fis[0].Open()
	}

	if !fis[0].IsDir() {
		// Pick the last file mount.
		return fis[len(fis)-1].Open()
	}

	openers := make([]func() (afero.File, error), len(fis))
	for i := len(fis) - 1; i >= 0; i-- {
		fi := fis[i]
		openers[i] = func() (afero.File, error) {
			return fi.Open()
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

func (rmfs *RootMappingFs) Mounts(base string) ([]fs.FileMetaInfo, error) {
	base = mapKey(cleanName(base))
	roots := rmfs.getRootsWithPrefix(base)

	if roots == nil {
		return nil, nil
	}

	fss := make([]fs.FileMetaInfo, len(roots))
	for i, r := range roots {
		fss[i] = r.ToFi
	}

	return fss, nil
}

func (rmfs *RootMappingFs) getRootsWithPrefix(prefix string) []RootMapping {
	var roots []RootMapping
	rmfs.rootMapToReal.WalkPrefix(prefix, func(b string, v any) bool {
		roots = append(roots, v.([]RootMapping)...)
		return false
	})

	return roots
}

func (rmfs *RootMappingFs) ReverseLookup(filename string) ([]ComponentPath, error) {
	return rmfs.ReverseLookupComponent("", filename)
}

func (rmfs *RootMappingFs) ReverseLookupComponent(component, filename string) ([]ComponentPath, error) {
	filename = cleanName(filename)
	key := mapKey(filename)

	longestPrefix, roots := rmfs.getRootsReverse(key)

	if len(roots) == 0 {
		return nil, nil
	}

	var cps []ComponentPath

	base := strings.TrimPrefix(key, longestPrefix)
	dir, name := filepath.Split(base)

	for _, first := range roots {
		if component != "" && first.FromBase != component {
			continue
		}

		cps = append(cps, ComponentPath{
			Component: first.FromBase,
			Path:      paths.ToSlashTrimLeading(paths.FilePathSeparator + filepath.Join(first.path(), dir, name)),
		})
	}

	return cps, nil
}

func (rmfs *RootMappingFs) getRootsReverse(key string) (string, []RootMapping) {
	tree := rmfs.realMapToRoot
	s, v, found := tree.LongestPrefix(key)
	if !found {
		return "", nil
	}
	return s, v.([]RootMapping)
}
