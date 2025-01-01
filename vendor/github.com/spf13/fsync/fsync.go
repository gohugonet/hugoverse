// Copyright (C) 2012 Mostafa Hajizadeh
// Copyright (C) 2014-2022 Steve Francia

// package fsync keeps two files or directories in sync.
//
//         err := fsync.Sync("~/dst", ".")
//
// After the above code, if err is nil, every file and directory in the current
// directory is copied to ~/dst and has the same permissions. Consequent calls
// will only copy changed or new files.
//
// SyncTo is a helper function which helps you sync a groups of files or
// directories into a single destination. For instance, calling
//
//     SyncTo("public", "build/app.js", "build/app.css", "images", "fonts")
//
// is equivalent to calling
//
//     Sync("public/app.js", "build/app.js")
//     Sync("public/app.css", "build/app.css")
//     Sync("public/images", "images")
//     Sync("public/fonts", "fonts")
//
// Actually, this is how SyncTo is implemented: consequent calls to Sync.
//
// By default, sync code ignores extra files in the destination that don’t have
// identicals in the source. Setting Delete field of a Syncer to true changes
// this behavior and deletes these extra files.

package fsync

import (
	"bytes"
	"errors"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"
)

var ErrFileOverDir = errors.New(
	"fsync: trying to overwrite a non-empty directory with a file")

// FileInfo contains the shared methods between os.FileInfo and fs.DirEntry.
type FileInfo interface {
	Name() string
	IsDir() bool
}

// Sync copies files and directories inside src into dst.
func Sync(dst, src string) error {
	return NewSyncer().Sync(dst, src)
}

// SyncTo syncs srcs files and directories into to directory.
func SyncTo(to string, srcs ...string) error {
	return NewSyncer().SyncTo(to, srcs...)
}

// Type Syncer provides functions for syncing files.
type Syncer struct {
	// Set this to true to delete files in the destination that don't exist
	// in the source.
	Delete bool
	// To allow certain files to remain in the destination, implement this function.
	// Return true to skip file, false to delete.
	// Note that src may be either os.FileInfo or fs.DirEntry depending on the file system.
	DeleteFilter func(f FileInfo) bool
	// By default, modification times are synced. This can be turned off by
	// setting this to true.
	NoTimes bool
	// NoChmod disables permission mode syncing.
	NoChmod bool
	// Implement this function to skip Chmod syncing for only certain files
	// or directories. Return true to skip Chmod.
	ChmodFilter func(dst, src os.FileInfo) bool

	// TODO add options for not checking content for equality

	SrcFs  afero.Fs
	DestFs afero.Fs
}

// NewSyncer creates a new instance of Syncer with default options.
func NewSyncer() *Syncer {
	s := Syncer{SrcFs: new(afero.OsFs), DestFs: new(afero.OsFs)}
	s.DeleteFilter = func(f FileInfo) bool {
		return false
	}
	return &s
}

// Sync copies files and directories inside src into dst.
func (s *Syncer) Sync(dst, src string) error {
	// make sure src exists
	if _, err := s.SrcFs.Stat(src); err != nil {
		return err
	}
	// return error instead of replacing a non-empty directory with a file
	if b, err := s.checkDir(dst, src); err != nil {
		return err
	} else if b {
		return ErrFileOverDir
	}

	return s.syncRecover(dst, src)
}

// SyncTo syncs srcs files or directories into to directory.
func (s *Syncer) SyncTo(to string, srcs ...string) error {
	for _, src := range srcs {
		dst := filepath.Join(to, filepath.Base(src))
		if err := s.Sync(dst, src); err != nil {
			return err
		}
	}
	return nil
}

// syncRecover handles errors and calls sync
func (s *Syncer) syncRecover(dst, src string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case runtime.Error:
				panic(r)
			case error:
				err = r
			default:
				panic(r)
			}
		}
	}()

	s.sync(dst, src)
	return nil
}

// sync updates dst to match with src, handling both files and directories.
func (s *Syncer) sync(dst, src string) {
	// sync permissions and modification times after handling content
	defer s.syncstats(dst, src)

	// read files info
	dstat, err := s.DestFs.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	sstat, err := s.SrcFs.Stat(src)
	if err != nil && os.IsNotExist(err) {
		return // src was deleted before we could copy it
	}
	check(err)

	if !sstat.IsDir() {
		// src is a file
		// delete dst if its a directory
		if dstat != nil && dstat.IsDir() {
			check(s.DestFs.RemoveAll(dst))
		}
		if !s.equal(dst, src, dstat, sstat) {
			// perform copy
			df, err := s.DestFs.Create(dst)
			check(err)
			defer df.Close()
			sf, err := s.SrcFs.Open(src)
			if os.IsNotExist(err) {
				return
			}
			check(err)
			defer sf.Close()
			_, err = io.Copy(df, sf)
			if os.IsNotExist(err) {
				return
			}
			check(err)
		}
		return
	}

	// src is a directory
	// make dst if necessary
	if dstat == nil {
		// dst does not exist; create directory
		check(s.DestFs.MkdirAll(dst, 0o755)) // permissions will be synced later
	} else if !dstat.IsDir() {
		// dst is a file; remove and create directory
		check(s.DestFs.Remove(dst))
		check(s.DestFs.MkdirAll(dst, 0o755)) // permissions will be synced later
	}

	// make a map of filenames for quick lookup; used in deletion
	// deletion below
	m := make(map[string]bool)
	err = withDirEntry(s.SrcFs, src, func(fi FileInfo) bool {
		dst2 := filepath.Join(dst, fi.Name())
		src2 := filepath.Join(src, fi.Name())
		s.sync(dst2, src2)
		m[fi.Name()] = true

		return false
	})

	if os.IsNotExist(err) {
		return
	}
	check(err)

	// delete files from dst that does not exist in src
	if s.Delete {
		err = withDirEntry(s.DestFs, dst, func(fi FileInfo) bool {
			if !m[fi.Name()] && !s.DeleteFilter(fi) {
				check(s.DestFs.RemoveAll(filepath.Join(dst, fi.Name())))
			}
			return false
		})
		check(err)

	}
}

// syncstats makes sure dst has the same pemissions and modification time as src
func (s *Syncer) syncstats(dst, src string) {
	// get file infos; return if not exist and panic if error
	dstat, err1 := s.DestFs.Stat(dst)
	sstat, err2 := s.SrcFs.Stat(src)
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		return
	}
	check(err1)
	check(err2)

	// update dst's permission bits
	noChmod := s.NoChmod
	if !noChmod && s.ChmodFilter != nil {
		noChmod = s.ChmodFilter(dstat, sstat)
	}
	if !noChmod {
		if dstat.Mode().Perm() != sstat.Mode().Perm() {
			check(s.DestFs.Chmod(dst, sstat.Mode().Perm()))
		}
	}

	// update dst's modification time
	if !s.NoTimes {
		if !dstat.ModTime().Equal(sstat.ModTime()) {
			err := s.DestFs.Chtimes(dst, sstat.ModTime(), sstat.ModTime())
			check(err)
		}
	}
}

// equal returns true if both dst and src files are equal
func (s *Syncer) equal(dst, src string, dstat, sstat os.FileInfo) bool {
	if sstat == nil || dstat == nil {
		return false
	}

	// check sizes
	if dstat.Size() != sstat.Size() {
		return false
	}

	// both have the same size, check the contents
	f1, err := s.DestFs.Open(dst)
	check(err)
	defer f1.Close()
	f2, err := s.SrcFs.Open(src)
	check(err)
	defer f2.Close()
	buf1 := make([]byte, 1000)
	buf2 := make([]byte, 1000)
	for {
		// read from both
		n1, err := f1.Read(buf1)
		if err != nil && err != io.EOF {
			panic(err)
		}
		n2, err := f2.Read(buf2)
		if err != nil && err != io.EOF {
			panic(err)
		}

		// compare read bytes
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false
		}

		// end of both files
		if n1 == 0 && n2 == 0 {
			break
		}
	}

	return true
}

// checkDir returns true if dst is a non-empty directory and src is a file
func (s *Syncer) checkDir(dst, src string) (b bool, err error) {
	// read file info
	dstat, err := s.DestFs.Stat(dst)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	sstat, err := s.SrcFs.Stat(src)
	if err != nil {
		return false, err
	}

	// return false is dst is not a directory or src is a directory
	if !dstat.IsDir() || sstat.IsDir() {
		return false, nil
	}

	// dst is a directory and src is a file
	// check if dst is non-empty
	// read dst directory
	var isNonEmpty bool
	err = withDirEntry(s.DestFs, dst, func(FileInfo) bool {
		isNonEmpty = true
		return true
	})

	return isNonEmpty, err
}

func withDirEntry(fs afero.Fs, path string, fn func(FileInfo) bool) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if rdf, ok := f.(iofs.ReadDirFile); ok {
		fis, err := rdf.ReadDir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			if fn(fi) {
				return nil
			}
		}
		return nil
	}

	fis, err := f.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		if fn(fi) {
			return nil
		}
	}

	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
