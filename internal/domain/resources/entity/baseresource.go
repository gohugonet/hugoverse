package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/identity"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/media"
	"io"
)

type baseResource interface {
	baseResourceResource
	baseResourceInternal
	stale.Staler
}

type baseResourceResource interface {
	resources.Cloner
	resources.Copier
	resources.ContentProvider
	resources.Resource
	resources.Identifier
}

type baseResourceInternal interface {
	resources.Source
	resources.NameNormalizedProvider

	fileInfo
	mediaTypeAssigner
	targetPather

	ReadSeekCloser() (pio.ReadSeekCloser, error)

	identity.IdentityGroupProvider
	identity.DependencyManagerProvider

	// For internal use.
	cloneWithUpdates(config *valueobject.TransformationUpdate) (baseResource, error)
	tryTransformedFileCache(key string, u *valueobject.TransformationUpdate) io.ReadCloser

	getResourcePaths() valueobject.ResourcePaths

	specProvider
	openPublishFileForWriting(relTargetPath string) (io.WriteCloser, error)
}

type specProvider interface {
	getSpec() *valueobject.Spec
}

type fileInfo interface {
	SetOpenSource(pio.OpenReadSeekCloser)
	SetSourceFilenameIsHash(bool)
	SetTargetPath(resources.ResourcePaths)
	Size() int64
	hashProvider
}

type hashProvider interface {
	Hash() string
}

// metaAssigner allows updating the media type in resources that supports it.
type mediaTypeAssigner interface {
	SetMediaType(mediaType media.Type)
}

type targetPather interface {
	TargetPath() string
}

type transformableResource interface {
	baseResourceInternal

	resources.ContentProvider
	resources.Resource
	resources.Identifier
	stale.Staler
	resources.Copier
}
