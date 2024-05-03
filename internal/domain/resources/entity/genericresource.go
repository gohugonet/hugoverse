package entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/identity"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"io"
	"strings"
	"sync"
)

// genericResource represents a generic linkable resource.
type genericResource struct {
	UrlService   resources.Url
	MediaService resources.MediaTypes

	publishInit *sync.Once

	publishFs afero.Fs

	sd    valueobject.ResourceSourceDescriptor
	paths valueobject.ResourcePaths

	resourceCache *ResourceCache

	sourceFilenameIsHash bool

	h *valueobject.ResourceHash // A hash of the source content. Is only calculated in caching situations.

	stale.Staler

	title  string
	name   string
	params map[string]any

	spec *valueobject.Spec
}

func (gr *genericResource) IdentifierBase() string {
	return gr.sd.Path.IdentifierBase()
}

func (gr *genericResource) GetIdentityGroup() identity.Identity {
	return gr.sd.GroupIdentity
}

func (gr *genericResource) GetDependencyManager() identity.Manager {
	return gr.sd.DependencyManager
}

func (gr *genericResource) ReadSeekCloser() (pio.ReadSeekCloser, error) {
	return gr.sd.OpenReadSeekCloser()
}

func (gr *genericResource) Clone() resources.Resource {
	return gr.clone()
}

func (gr *genericResource) Size() int64 {
	gr.Hash()
	return gr.h.Size
}

func (gr *genericResource) Hash() string {
	if err := gr.h.Setup(gr); err != nil {
		panic(err)
	}
	return gr.h.Value
}

func (gr *genericResource) SetOpenSource(openSource pio.OpenReadSeekCloser) {
	gr.sd.OpenReadSeekCloser = openSource
}

func (gr *genericResource) SetSourceFilenameIsHash(b bool) {
	gr.sourceFilenameIsHash = b
}

func (gr *genericResource) SetTargetPath(d resources.ResourcePaths) {
	gr.paths = valueobject.ResourcePaths{
		Dir:             d.PathDir(),
		BaseDirTarget:   d.PathBaseDirTarget(),
		BaseDirLink:     d.PathBaseDirLink(),
		TargetBasePaths: d.PathTargetBasePaths(),
		File:            d.PathFile(),
	}
}

func (gr *genericResource) CloneTo(targetPath string) resources.Resource {
	c := gr.clone()
	c.paths = c.paths.FromTargetPath(targetPath)
	return c
}

func (gr *genericResource) Content(context.Context) (any, error) {
	r, err := gr.ReadSeekCloser()
	if err != nil {
		return "", err
	}
	defer r.Close()

	return pio.ReadString(r)
}

func (gr *genericResource) Err() resources.ResourceError {
	return nil
}

func (gr *genericResource) Data() any {
	return gr.sd.Data
}

func (gr *genericResource) Key() string {
	basePath := gr.UrlService.BasePathNoSlash()
	var key string
	if basePath == "" {
		key = gr.RelPermalink()
	} else {
		key = strings.TrimPrefix(gr.RelPermalink(), basePath)
	}

	return key
}

func (gr *genericResource) MediaType() media.Type {
	return gr.sd.MediaType
}

func (gr *genericResource) SetMediaType(mediaType media.Type) {
	gr.sd.MediaType = mediaType
}

func (gr *genericResource) Name() string {
	return gr.name
}

func (gr *genericResource) NameNormalized() string {
	return gr.sd.NameNormalized
}

func (gr *genericResource) Params() maps.Params {
	return gr.params
}

func (gr *genericResource) Publish() error {
	var err error
	gr.publishInit.Do(func() {
		targetFilenames := gr.getResourcePaths().TargetFilenames()

		if gr.sourceFilenameIsHash {
			// This is a processed images. We want to avoid copying it if it hasn't changed.
			var changedFilenames []string
			for _, targetFilename := range targetFilenames {
				if _, err := gr.publishFs.Stat(targetFilename); err == nil {
					continue
				}
				changedFilenames = append(changedFilenames, targetFilename)
			}
			if len(changedFilenames) == 0 {
				return
			}
			targetFilenames = changedFilenames
		}
		var fr pio.ReadSeekCloser
		fr, err = gr.ReadSeekCloser()
		if err != nil {
			return
		}
		defer fr.Close()

		var fw io.WriteCloser
		fw, err = helpers.OpenFilesForWriting(gr.publishFs, targetFilenames...)
		if err != nil {
			return
		}
		defer fw.Close()

		_, err = io.Copy(fw, fr)
	})

	return err
}

func (gr *genericResource) RelPermalink() string {
	return gr.UrlService.BasePathNoSlash() + paths.PathEscape(gr.paths.TargetLink())
}

func (gr *genericResource) Permalink() string {
	return gr.UrlService.BasePathNoSlash() + paths.PathEscape(gr.paths.TargetPath())
}

func (gr *genericResource) ResourceType() string {
	return gr.MediaType().MainType
}

func (gr *genericResource) String() string {
	return fmt.Sprintf("Resource(%s: %s)", gr.ResourceType(), gr.name)
}

// Path is stored with Unix style slashes.
func (gr *genericResource) TargetPath() string {
	return gr.paths.TargetPath()
}

func (gr *genericResource) Title() string {
	return gr.title
}

func (gr *genericResource) getResourcePaths() valueobject.ResourcePaths {
	return gr.paths
}

func (gr *genericResource) tryTransformedFileCache(key string, u *valueobject.TransformationUpdate) io.ReadCloser {
	fi, f, meta, found := gr.resourceCache.GetFromFile(key)
	if !found {
		return nil
	}
	u.SourceFilename = &fi.Name
	mt, _ := gr.MediaService.LookByType(meta.MediaTypeV)
	u.MediaType = mt
	u.Data = meta.MetaData
	u.TargetPath = meta.Target
	return f
}

func (gr *genericResource) mergeData(in map[string]any) {
	if len(in) == 0 {
		return
	}
	if gr.sd.Data == nil {
		gr.sd.Data = make(map[string]any)
	}
	for k, v := range in {
		if _, found := gr.sd.Data[k]; !found {
			gr.sd.Data[k] = v
		}
	}
}

func (gr *genericResource) cloneWithUpdates(u *valueobject.TransformationUpdate) (baseResource, error) {
	r := gr.clone()

	if u.Content != nil {
		r.sd.OpenReadSeekCloser = func() (pio.ReadSeekCloser, error) {
			return pio.NewReadSeekerNoOpCloserFromString(*u.Content), nil
		}
	}

	r.sd.MediaType = u.MediaType

	if u.SourceFilename != nil {
		if u.SourceFs == nil {
			return nil, errors.New("sourceFs is nil")
		}
		r.SetOpenSource(func() (pio.ReadSeekCloser, error) {
			return u.SourceFs.Open(*u.SourceFilename)
		})
	} else if u.SourceFs != nil {
		return nil, errors.New("sourceFs is set without sourceFilename")
	}

	if u.TargetPath == "" {
		return nil, errors.New("missing targetPath")
	}

	r.SetTargetPath(r.paths.FromTargetPath(u.TargetPath))
	r.mergeData(u.Data)

	return r, nil
}

func (gr *genericResource) clone() *genericResource {
	clone := *gr
	clone.publishInit = &sync.Once{}
	return &clone
}

func (gr *genericResource) openPublishFileForWriting(relTargetPath string) (io.WriteCloser, error) {
	filenames := gr.paths.FromTargetPath(relTargetPath).TargetFilenames()
	return helpers.OpenFilesForWriting(gr.publishFs, filenames...)
}

func (gr *genericResource) getSpec() *valueobject.Spec {
	return gr.spec
}
