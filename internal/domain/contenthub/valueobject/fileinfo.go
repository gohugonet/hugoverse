package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path/filepath"
	"sync"
)

// File describes a source file.
type File struct {
	fim fs.FileMetaInfo

	uniqueID string
	lazyInit sync.Once
}

func NewFileInfo(fi fs.FileMetaInfo) *File {
	return &File{
		fim: fi,
	}
}

// Filename returns a file's absolute path and filename on disk.
func (fi *File) Filename() string { return fi.fim.FileName() }

// Path gets the relative path including file name and extension.  The directory
// is relative to the content root.
func (fi *File) Path() string { return filepath.Join(fi.p().Dir()[1:], fi.p().Name()) }

// Dir gets the name of the directory that contains this file.  The directory is
// relative to the content root.
func (fi *File) Dir() string {
	return fi.pathToDir(fi.p().Dir())
}

// Extension is an alias to Ext().
// Deprecated: Use Ext() instead.
func (fi *File) Extension() string {
	hugo.Deprecate(".File.Extension", "Use .File.Ext instead.", "v0.96.0")
	return fi.Ext()
}

// Ext returns a file's extension without the leading period (e.g. "md").
func (fi *File) Ext() string { return fi.p().Ext() }

// LogicalName returns a file's name and extension (e.g. "page.sv.md").
func (fi *File) LogicalName() string {
	return fi.p().Name()
}

// BaseFileName returns a file's name without extension (e.g. "page.sv").
func (fi *File) BaseFileName() string {
	return fi.p().NameNoExt()
}

// TranslationBaseName returns a file's translation base name without the
// language segment (e.g. "page").
func (fi *File) TranslationBaseName() string { return fi.p().NameNoIdentifier() }

// ContentBaseName is a either TranslationBaseName or name of containing folder
// if file is a bundle.
func (fi *File) ContentBaseName() string {
	return fi.p().BaseNameNoIdentifier()
}

// Section returns a file's section.
func (fi *File) Section() string {
	return fi.p().Section()
}

// UniqueID returns a file's unique, MD5 hash identifier.
func (fi *File) UniqueID() string {
	fi.init()
	return fi.uniqueID
}

// FileInfo returns a file's underlying os.FileInfo.
func (fi *File) FileInfo() fs.FileMetaInfo { return fi.fim }

func (fi *File) String() string { return fi.BaseFileName() }

// Open implements ReadableFile.
func (fi *File) Open() (io.ReadSeekCloser, error) {
	f, err := fi.fim.Open()

	return f, err
}

func (fi *File) IsZero() bool {
	return fi == nil
}

// We create a lot of these FileInfo objects, but there are parts of it used only
// in some cases that is slightly expensive to construct.
func (fi *File) init() {
	fi.lazyInit.Do(func() {
		fi.uniqueID = helpers.MD5String(filepath.ToSlash(fi.Path()))
	})
}

func (fi *File) pathToDir(s string) string {
	if s == "" {
		return s
	}
	return filepath.FromSlash(s[1:] + "/")
}

func (fi *File) p() *paths.Path {
	return fi.fim.Path().Unnormalized()
}
