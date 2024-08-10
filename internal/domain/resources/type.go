package resources

import (
	"context"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/hexec"
	"github.com/gohugonet/hugoverse/pkg/images/exif"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/spf13/afero"
	"github.com/tdewolff/minify/v2"
	"image/color"
	"time"
)

type Workspace interface {
	Fs

	ImageConfig
	CacheConfig
	MediaTypesConfig
	SecurityConfig
	OutputFormatsConfig
	MinifyConfig
}

type OutputFormatsConfig interface {
	AllOutputFormats() output.Formats
}

type MinifyConfig interface {
	IsMinifyPublish() bool
	GetMinifier(s string) minify.Minifier
	Minifiers(mediaTypes media.Types, cb func(media.Type, minify.Minifier))
}

type SecurityConfig interface {
	ExecAuth() hexec.ExecAuth
}

type Fs interface {
	SourceFs() afero.Fs
	AssetsFs() afero.Fs
	PublishFs() afero.Fs
	ResourcesCacheFs() afero.Fs

	AssetsFsRealFilename(rel string) string
	AssetsFsRealDirs(from string) []string
	AssetsFsMakePathRelative(filename string, checkExists bool) (string, bool)

	NewBasePathFs(source afero.Fs, path string) afero.Fs
	Glob(fs afero.Fs, pattern string, handle func(fi fs.FileMetaInfo) (bool, error)) error
}

type CacheConfig interface {
	CachesIterator(func(cacheKey string, isResourceDir bool, dir string, age time.Duration))
}

type ImageConfig interface {
	ExifDecoder() (*exif.Decoder, error)
	ImageHint() webpoptions.EncodingPreset
	ImageQuality() int
	Resampling() gift.Resampling
	ResamplingStr() string
	Anchor() gift.Anchor
	AnchorStr() string
	BgColor() color.Color
	BgColorStr() string
	SourceHash() string
}

type MediaTypesConfig interface {
	LookFirstBySuffix(suffix string) (media.Type, media.SuffixInfo, bool)
	LookByType(mediaType string) (media.Type, bool)
	MediaTypes() media.Types
}

type SiteUrl interface {
	BasePathNoSlash() string
}

type Resources interface {
	GetResource(pathname string) (Resource, error)
	GetResourceWithOpener(pathname string, opener io.OpenReadSeekCloser) (Resource, error)
}

type Resource interface {
	ReadSeekCloser() (io.ReadSeekCloser, error)
	TargetPath() string
}

type ResourceCopier interface {
	CloneTo(targetPath string) Resource
}

const (
	PostProcessPrefix = "__h_pp_l1"

	// The suffix has an '=' in it to prevent the minifier to remove any enclosing
	// quoutes around the attribute values.
	// See issue #8884.
	PostProcessSuffix = "__e="
)

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

// OriginProvider provides the original Resource if this is wrapped.
// This is an internal Hugo interface and not meant for use in the templates.
type OriginProvider interface {
	Origin() Resource
	GetFieldString(pattern string) (string, bool)
}

type TypeProvider interface {
	// ResourceType is the resource type. For most file types, this is the main
	// part of the MIME type, e.g. "images", "application", "text" etc.
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
	// So, for the images "/some/path/sunset.jpg" this will be "sunset.jpg".
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

// Identifier identifies a resource.
type Identifier interface {
	// Key is is mostly for internal use and should be considered opaque.
	// This value may change between Hugo versions.
	Key() string
}

// Copier is for internal use.
type Copier interface {
	CloneTo(targetPath string) Resource
}

type NameNormalizedProvider interface {
	// NameNormalized is the normalized name of this resource.
	// For internal use (for now).
	NameNormalized() string
}

type ResourcePaths interface {
	PathDir() string
	PathBaseDirTarget() string
	PathBaseDirLink() string
	PathTargetBasePaths() []string
	PathFile() string
}

type Cloner interface {
	Clone() Resource
}

// ImageResource represents an images resource.
type ImageResource interface {
	Resource
	ImageResourceOps
}

// MetaProvider provides metadata about a resource.
type MetaProvider interface {
	NameTitleProvider
	ParamsProvider
}

// ReadSeekCloserResource is a Resource that supports loading its content.
type ReadSeekCloserResource interface {
	MediaType() media.Type
	io.ReadSeekCloserProvider
}
