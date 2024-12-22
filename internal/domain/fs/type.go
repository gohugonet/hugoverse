package fs

import (
	"github.com/gohugonet/hugoverse/pkg/media"
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

type ComponentPath interface {
	GetComponent() string
	GetPath() string
	GetLang() string
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
	RelativeFilename() (string, error)
	Component() string
	Root() string
}

type (
	WalkFunc func(path string, info FileMetaInfo) error
	WalkHook func(dir FileMetaInfo, path string, readdir []FileMetaInfo) ([]FileMetaInfo, error)
)

type WalkCallback struct {
	// Will be called in order.
	HookPre  WalkHook // Optional.
	WalkFn   WalkFunc
	HookPost WalkHook // Optional.
}

type WalkwayConfig struct {
	// One or both of these may be pre-set.
	Info       FileMetaInfo               // The start info.
	DirEntries []FileMetaInfo             // The start info's dir entries.
	IgnoreFile func(filename string) bool // Optional

	// Some optional flags.
	FailOnNotExist bool // If set, return an error if a directory is not found.
	SortDirEntries bool // If set, sort the dir entries by Name before calling the WalkFn, default is ReaDir order.
}
