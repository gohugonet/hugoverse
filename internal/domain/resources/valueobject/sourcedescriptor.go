package valueobject

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"mime"
)

type ResourceSourceDescriptor struct {
	// The source Content.
	OpenReadSeekCloser io.OpenReadSeekCloser

	// The canonical source path.
	Path *paths.Path

	// The normalized name of the resource.
	NameNormalized string

	// The name of the resource as it was read from the source.
	NameOriginal string

	// Any base paths prepended to the target path. This will also typically be the
	// language code, but setting it here means that it should not have any effect on
	// the permalink.
	// This may be several values. In multihost mode we may publish the same resources to
	// multiple targets.
	TargetBasePaths []string

	TargetPath           string
	BasePathRelPermalink string
	BasePathTargetPath   string

	// The Data to associate with this resource.
	Data map[string]any

	// Delay publishing until either Permalink or RelPermalink is called. Maybe never.
	LazyPublish bool

	// Set when its known up front, else it's resolved from the target filename.
	MediaType media.Type

	// Used to track dependencies (e.g. imports). May be nil if that's of no concern.
	DependencyManager identity.Manager

	// A shared identity for this resource and all its clones.
	// If this is not set, an Identity is created.
	GroupIdentity identity.Identity
}

func NewResourceSourceDescriptor(
	pathname string,
	path *paths.Path,
	mediaService resources.MediaTypes,
	openReadSeekCloser io.OpenReadSeekCloser) (*ResourceSourceDescriptor, error) {

	sd := &ResourceSourceDescriptor{
		LazyPublish:        true,
		OpenReadSeekCloser: openReadSeekCloser,
		Path:               path,
		GroupIdentity:      path,
		TargetPath:         pathname,
		Data:               make(map[string]any),
	}

	if err := sd.setup(mediaService); err != nil {
		return nil, err
	}

	return sd, nil
}

func (fd *ResourceSourceDescriptor) setup(mediaService resources.MediaTypes) error {
	if fd.OpenReadSeekCloser == nil {
		panic(errors.New("OpenReadSeekCloser is nil"))
	}

	if fd.TargetPath == "" {
		panic(errors.New("RelPath is empty"))
	}

	if fd.Path == nil {
		fd.Path = paths.Parse("", fd.TargetPath)
	}

	if fd.TargetPath == "" {
		fd.TargetPath = fd.Path.Path()
	} else {
		fd.TargetPath = paths.ToSlashPreserveLeading(fd.TargetPath)
	}

	fd.BasePathRelPermalink = paths.ToSlashPreserveLeading(fd.BasePathRelPermalink)
	if fd.BasePathRelPermalink == "/" {
		fd.BasePathRelPermalink = ""
	}
	fd.BasePathTargetPath = paths.ToSlashPreserveLeading(fd.BasePathTargetPath)
	if fd.BasePathTargetPath == "/" {
		fd.BasePathTargetPath = ""
	}

	fd.TargetPath = paths.ToSlashPreserveLeading(fd.TargetPath)
	for i, base := range fd.TargetBasePaths {
		dir := paths.ToSlashPreserveLeading(base)
		if dir == "/" {
			dir = ""
		}
		fd.TargetBasePaths[i] = dir
	}

	if fd.NameNormalized == "" {
		fd.NameNormalized = fd.TargetPath
	}

	if fd.NameOriginal == "" {
		fd.NameOriginal = fd.NameNormalized
	}

	mediaType := fd.MediaType
	if mediaType.IsZero() {
		ext := fd.Path.Ext()
		var (
			found      bool
			suffixInfo media.SuffixInfo
		)
		mediaType, suffixInfo, found = mediaService.LookFirstBySuffix(ext)
		// TODO(bep) we need to handle these ambiguous types better, but in this context
		// we most likely want the application/xml type.
		if suffixInfo.Suffix == "xml" && mediaType.SubType == "rss" {
			mediaType, found = mediaService.LookByType("application/xml")
		}

		if !found {
			// A fallback. Note that mime.TypeByExtension is slow by Hugo standards,
			// so we should configure media types to avoid this lookup for most
			// situations.
			mimeStr := mime.TypeByExtension("." + ext)
			if mimeStr != "" {
				mediaType, _ = media.FromStringAndExt(mimeStr, ext)
			}
		}
	}

	fd.MediaType = mediaType

	if fd.DependencyManager == nil {
		fd.DependencyManager = identity.NewManager("resource")
	}

	return nil
}
