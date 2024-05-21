package valueobject

import (
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

		addMapping(paths.FilePathSeparator+rm.From, rm, rootMapToReal)
		rev := rm.To
		if !strings.HasPrefix(rev, paths.FilePathSeparator) {
			rev = paths.FilePathSeparator + rev
		}

		addMapping(rev, rm, realMapToRoot)

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

func (fs *RootMappingFs) UnwrapFilesystem() afero.Fs {
	return fs.Fs
}

// Stat returns the os.FileInfo structure describing a given file.  If there is
// an error, it will be of type *os.PathError.
func (fs *RootMappingFs) Stat(name string) (os.FileInfo, error) {
	fis, err := fs.doStat(name)
	if err != nil {
		return nil, err
	}
	// first win
	return fis[0], nil
}

func (fs *RootMappingFs) doStat(name string) ([]FileMetaInfo, error) {
	fis, err := fs.doDoStat(name)
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

func (fs *RootMappingFs) doDoStat(name string) ([]FileMetaInfo, error) {
	name = cleanName(name)
	key := paths.FilePathSeparator + name

	roots := fs.getRoot(key)

	if roots == nil {
		if fs.hasPrefix(key) {
			// We have directories mounted below this.
			// Make it look like a directory.
			panic("single file mount not supported yet")
		}

		// Find any real directories with this key.
		_, roots := fs.getRoots(key)
		if roots == nil {
			return nil, &os.PathError{Op: "LStat", Path: name, Err: os.ErrNotExist}
		}

		var err error
		var fis []FileMetaInfo

		for _, rm := range roots {
			var fi FileMetaInfo
			fi, err = fs.statRoot(rm, name)
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

	return []FileMetaInfo{newDirNameOnlyFileInfo(name, roots[0].Meta, fs.virtualDirOpener(name))}, nil
}

func (fs *RootMappingFs) statRoot(root RootMapping, filename string) (*FileInfo, error) {
	dir, name := filepath.Split(filename)

	filename = root.filename(filename)
	fi, err := fs.Fs.Stat(filename)
	if err != nil {
		return nil, err
	}

	var opener func() (afero.File, error)
	if !fi.IsDir() {
		// Open the file directly.
		// Opens the real file directly.
		opener = func() (afero.File, error) {
			return fs.Fs.Open(filename)
		}
	} else {
		// Make sure metadata gets applied in ReadDir.
		opener = func() (afero.File, error) {
			f, err := fs.Fs.Open(name)
			if err != nil {
				return nil, err
			}

			df := NewDirFileWithOpener(f, fs.collectDirEntries)
			return df, nil
		}
	}

	fim := NewFileInfoWithOpener(fi, name, opener)

	return fim, nil
}

func (fs *RootMappingFs) getRoot(key string) []RootMapping {
	v, found := fs.rootMapToReal.Get(key)
	if !found {
		return nil
	}

	return v.([]RootMapping)
}

func (fs *RootMappingFs) hasPrefix(prefix string) bool {
	hasPrefix := false
	fs.rootMapToReal.WalkPrefix(prefix, func(b string, v any) bool {
		hasPrefix = true
		return true
	})

	return hasPrefix
}

func (fs *RootMappingFs) getRoots(key string) (string, []RootMapping) {
	tree := fs.rootMapToReal
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

func (fs *RootMappingFs) collectDirEntries(prefix string) ([]iofs.DirEntry, error) {
	prefix = paths.FilePathSeparator + cleanName(prefix)

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
					return fs.Open(filepath.Join(rm.From, name))
				}
				fi = newDirNameOnlyFileInfo(name, meta, opener)
			}
			fis = append(fis, fi)
		}

		f.Close()

		return nil
	}

	// First add any real files/directories.
	rms := fs.getRoot(prefix)
	for _, rm := range rms {
		if err := collectDir(rm, rm.ToFi); err != nil {
			return nil, err
		}
	}

	// Next add any file mounts inside the given directory.
	prefixInside := prefix + filepathSeparator
	fs.rootMapToReal.WalkPrefix(prefixInside, func(s string, v any) bool {
		if (strings.Count(s, filepathSeparator) - level) != 1 {
			// This directory is not part of the current, but we
			// need to include the first name part to make it
			// navigable.
			path := strings.TrimPrefix(s, prefixInside)
			parts := strings.Split(path, filepathSeparator)
			name := parts[0]

			if seen[name] {
				return false
			}
			seen[name] = true
			opener := func() (afero.File, error) {
				return fs.Open(path)
			}

			fi := newDirNameOnlyFileInfo(name, nil, opener)
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
				return fs.Open(rm.From)
			}
			fi := newDirNameOnlyFileInfo(name, rm.Meta, opener)
			fis = append(fis, fi)
		}

		return false
	})

	// Finally add any ancestor dirs with files in this directory.
	ancestors := fs.getAncestors(prefix)
	for _, root := range ancestors {
		subdir := strings.TrimPrefix(prefix, root.key)
		for _, rm := range root.roots {
			if rm.fi.IsDir() {
				fi, err := rm.fi.Meta().JoinStat(subdir)
				if err == nil {
					if err := collectDir(rm, fi); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return fis, nil
}
