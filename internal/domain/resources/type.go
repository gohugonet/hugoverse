package resources

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/hexec"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/image/exif"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/spf13/afero"
	"go.opencensus.io/resource"
	"time"
)

type Workspace interface {
	SourceFs() afero.Fs
	ResourcesCacheFs() afero.Fs
	NewBasePathFs(source afero.Fs, path string) afero.Fs

	ExecAuth() hexec.ExecAuth

	ExifDecoder() (*exif.Decoder, error)

	CachesIterator(func(cacheKey string, isResourceDir bool, dir string, age time.Duration) error)
}

type Resources interface {
	Creator
}

type Creator interface {
	GetResource(pathname string) (Resource, error)
}

type Resource interface {
	TypeProvider
	MediaTypeProvider
	LinksProvider
	NameTitleProvider
	ParamsProvider
	DataProvider
	ErrProvider
}

type PostPublishedResource interface {
	TypeProvider
	LinksProvider
	NameTitleProvider
	ParamsProvider
	DataProvider
	OriginProvider

	MediaType() map[string]any
}

// Source is an internal template and not meant for use in the templates. It
// may change without notice.
type Source interface {
	Publish() error
}

type TransformableResource interface {
	baseResourceInternal

	ContentProvider
	Resource
	Identifier
	stale.Staler
	resourceCopier
}

type baseResourceInternal interface {
	Source
	NameNormalizedProvider

	fileInfo
	mediaTypeAssigner
	targetPather

	ReadSeekCloser() (pio.ReadSeekCloser, error)

	identity.IdentityGroupProvider
	identity.DependencyManagerProvider
}

// metaAssigner allows updating the media type in resources that supports it.
type mediaTypeAssigner interface {
	setMediaType(mediaType media.Type)
}

type targetPather interface {
	TargetPath() string
}

// OriginProvider provides the original Resource if this is wrapped.
// This is an internal Hugo interface and not meant for use in the templates.
type OriginProvider interface {
	Origin() Resource
	GetFieldString(pattern string) (string, bool)
}

type TypeProvider interface {
	// ResourceType is the resource type. For most file types, this is the main
	// part of the MIME type, e.g. "image", "application", "text" etc.
	// For content pages, this value is "page".
	ResourceType() string
}

type MediaTypeProvider interface {
	// MediaType is this resource's MIME type.
	MediaType() media.Type
}

type LinksProvider interface {
	// Permalink represents the absolute link to this resource.
	Permalink() string

	// RelPermalink represents the host relative link to this resource.
	RelPermalink() string
}

type NameTitleProvider interface {
	// Name is the logical name of this resource. This can be set in the front matter
	// metadata for this resource. If not set, Hugo will assign a value.
	// This will in most cases be the base filename.
	// So, for the image "/some/path/sunset.jpg" this will be "sunset.jpg".
	// The value returned by this method will be used in the GetByPrefix and ByPrefix methods
	// on Resources.
	Name() string

	// Title returns the title if set in front matter. For content pages, this will be the expected value.
	Title() string
}

type ParamsProvider interface {
	// Params set in front matter for this resource.
	Params() maps.Params
}

type DataProvider interface {
	// Resource specific data set by Hugo.
	// One example would be .Data.Integrity for fingerprinted resources.
	Data() any
}

// ErrProvider provides an Err.
type ErrProvider interface {
	// Err returns an error if this resource is in an error state.
	// This will currently only be set for resources obtained from resources.GetRemote.
	Err() ResourceError
}

// ResourceError is the error return from .Err in Resource in error situations.
type ResourceError interface {
	error
	DataProvider
}

// ContentProvider provides Content.
// This should be used with care, as it will read the file content into memory, but it
// should be cached as effectively as possible by the implementation.
type ContentProvider interface {
	// Content returns this resource's content. It will be equivalent to reading the content
	// that RelPermalink points to in the published folder.
	// The return type will be contextual, and should be what you would expect:
	// * Page: template.HTML
	// * JSON: String
	// * Etc.
	Content(context.Context) (any, error)
}

// Identifier identifies a resource.
type Identifier interface {
	// Key is is mostly for internal use and should be considered opaque.
	// This value may change between Hugo versions.
	Key() string
}

// resourceCopier is for internal use.
type resourceCopier interface {
	cloneTo(targetPath string) resource.Resource
}

type NameNormalizedProvider interface {
	// NameNormalized is the normalized name of this resource.
	// For internal use (for now).
	NameNormalized() string
}

type fileInfo interface {
	setOpenSource(pio.OpenReadSeekCloser)
	setSourceFilenameIsHash(bool)
	setTargetPath(ResourcePaths)
	size() int64
	hashProvider
}

type hashProvider interface {
	hash() string
}

type ResourcePaths interface {
	// Dir This is the directory component for the target file or link.
	Dir() string

	// BaseDirTarget Any base directory for the target file. Will be prepended to Dir.
	BaseDirTarget() string

	// BaseDirLink This is the directory component for the link will be prepended to Dir.
	BaseDirLink() string

	// TargetBasePaths Set when publishing in a multihost setup.
	TargetBasePaths() []string

	// File This is the File component, e.g. "data.json".
	File() string
}
