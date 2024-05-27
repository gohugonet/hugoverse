package fs

import (
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"io/fs"
	"path/filepath"
)

type Dir interface {
	WorkingDir() string
	PublishDir() string
	ResourceDir() string

	MediaTypes() media.Types
}

type Fs interface {
	OriginFs
	PathFs
}

type PathFs interface {
	LayoutFs() afero.Fs
	ContentFs() afero.Fs
}

// OriginFs holds the core filesystems used by Hugo.
type OriginFs interface {
	// Origin is Hugo's source file system.
	// Note that this will always be a "plain" Afero filesystem:
	// * afero.OsFs when running in production
	// * afero.MemMapFs for many of the tests.
	Origin() afero.Fs

	// Publish is where Hugo publishes its rendered content.
	// It's mounted inside publishDir (default /public).
	Publish() afero.Fs

	// Working is a read-only file system
	// restricted to the project working dir.
	Working() afero.Fs
}

var FilepathSeparator = string(filepath.Separator)

// FilesystemUnwrapper returns the underlying filesystem.
type FilesystemUnwrapper interface {
	UnwrapFilesystem() afero.Fs
}

// FilesystemsUnwrapper returns the underlying filesystems.
type FilesystemsUnwrapper interface {
	UnwrapFilesystems() []afero.Fs
}

type FileMetaInfo interface {
	fs.FileInfo
	FileMeta
}

type FileMeta interface {
	Open() (afero.File, error)
	FileName() string

	Path() *paths.Path
	SetPath(path *paths.Path)
}

type (
	WalkFunc func(path string, info FileMetaInfo) error
	WalkHook func(dir FileMetaInfo, path string, readdir []FileMetaInfo) ([]FileMetaInfo, error)
)

type WalkCallback interface {
	PreHook() WalkHook // Optional.
	WalkHook() WalkFunc
	PostHook() WalkHook // Optional.
}
