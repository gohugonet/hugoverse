package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	"path/filepath"
	"sync"
)

// File describes a source file.
type File struct {
	fs.FileMetaInfo

	BundleType
	path *paths.Path

	uniqueID string
	lazyInit sync.Once
}

func NewFileInfo(fi fs.FileMetaInfo) (*File, error) {
	relName, err := fi.RelativeFilename()
	if err != nil {
		return nil, err
	}

	f := &File{
		FileMetaInfo: fi,

		path:       paths.Parse(files.ComponentFolderContent, relName),
		BundleType: BundleTypeFile,
	}

	isContent := files.IsContentExt(f.path.Ext())

	if isContent {
		switch f.path.NameNoIdentifier() {
		case "index":
			f.BundleType = BundleTypeLeaf
		case "_index":
			f.BundleType = BundleTypeBranch
		default:
			f.BundleType = BundleTypeContentSingle
		}
	}

	return f, nil
}

func (fi *File) PageFile() contenthub.File {
	return fi
}

func (fi *File) ShiftToResource() {
	if fi.IsContent() {
		fi.BundleType = BundleTypeContentResource
	}
}

// Filename returns a file's absolute path and filename on disk.
func (fi *File) Filename() string { return fi.FileMetaInfo.FileName() }

// RelPath Paths gets the relative path including file name and extension.  The directory
// is relative to the content root.
func (fi *File) RelPath() string { return filepath.Join(fi.p().Dir()[1:], fi.p().Name()) }

// Dir gets the name of the directory that contains this file.  The directory is
// relative to the content root.
func (fi *File) Dir() string {
	return fi.pathToDir(fi.p().Dir())
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

func (fi *File) Type() string {
	if sect := fi.Section(); sect != "" {
		return sect
	}
	return "page"
}

// UniqueID returns a file's unique, MD5 hash identifier.
func (fi *File) UniqueID() string {
	fi.init()
	return fi.uniqueID
}

// FileInfo returns a file's underlying os.FileInfo.
func (fi *File) FileInfo() fs.FileMetaInfo { return fi.FileMetaInfo }

func (fi *File) String() string { return fi.BaseFileName() }

// Open implements ReadableFile.
func (fi *File) Open() (io.ReadSeekCloser, error) {
	f, err := fi.FileMetaInfo.Open()

	return f, err
}

func (fi *File) Opener() io.OpenReadSeekCloser {
	return func() (io.ReadSeekCloser, error) {
		return fi.Open()
	}
}

func (fi *File) IsZero() bool {
	return fi == nil
}

// We create a lot of these FileInfo objects, but there are parts of it used only
// in some cases that is slightly expensive to construct.
func (fi *File) init() {
	fi.lazyInit.Do(func() {
		fi.uniqueID = helpers.MD5String(filepath.ToSlash(fi.RelPath()))
	})
}

func (fi *File) pathToDir(s string) string {
	if s == "" {
		return s
	}
	return filepath.FromSlash(s[1:] + "/")
}

func (fi *File) p() *paths.Path {
	return fi.path
}

func (fi *File) Paths() *paths.Path {
	return fi.p()
}

func (fi *File) Path() string {
	return fi.p().Path()
}
