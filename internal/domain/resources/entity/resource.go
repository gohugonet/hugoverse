package entity

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"io"
	"strings"
)

var (
	_ resources.Resource               = (*Resource)(nil)
	_ resources.ReadSeekCloserResource = (*Resource)(nil)
	_ hashProvider                     = (*Resource)(nil)
)

type Resource struct {
	stale.Staler
	h *valueobject.ResourceHash // A hash of the source content. Is only calculated in caching situations.

	title  string
	name   string
	params map[string]any

	sd    valueobject.ResourceSourceDescriptor
	paths valueobject.ResourcePaths

	publishFs afero.Fs
}

func (l *Resource) MediaType() media.Type {
	return l.sd.MediaType
}

func (l *Resource) ResourceType() string {
	return l.MediaType().MainType
}

func (l *Resource) RelPermalink() string {
	return l.paths.BasePathNoTrailingSlash + paths.PathEscape(l.paths.TargetLink())
}

func (l *Resource) Permalink() string {
	return l.paths.BasePathNoTrailingSlash + paths.PathEscape(l.paths.TargetPath())
}

func (l *Resource) Name() string {
	return l.name
}

func (l *Resource) Title() string {
	return l.title
}

func (l *Resource) Params() maps.Params {
	return l.params
}

func (l *Resource) Data() any {
	return l.sd.Data
}

func (l *Resource) Err() resources.ResourceError {
	return nil
}

func (l *Resource) ReadSeekCloser() (pio.ReadSeekCloser, error) {
	return l.sd.OpenReadSeekCloser()
}

func (l *Resource) Hash() string {
	if err := l.h.Setup(l); err != nil {
		panic(err)
	}
	return l.h.Value
}

func (l *Resource) Size() int64 {
	l.Hash()
	return l.h.Size
}

func (l *Resource) CloneTo(targetPath string) resources.Resource {
	c := l.clone()
	c.paths = c.paths.FromTargetPath(targetPath)
	return c
}

func (l *Resource) clone() *Resource {
	clone := *l
	return &clone
}

func (l *Resource) Key() string {
	basePath := l.paths.BasePathNoTrailingSlash
	var key string
	if basePath == "" {
		key = l.RelPermalink()
	} else {
		key = strings.TrimPrefix(l.RelPermalink(), basePath)
	}

	return key
}

func (l *Resource) openPublishFileForWriting(relTargetPath string) (io.WriteCloser, error) {
	filenames := l.paths.FromTargetPath(relTargetPath).TargetFilenames()
	return helpers.OpenFilesForWriting(l.publishFs, filenames...)
}

func (l *Resource) cloneWithUpdates(u *valueobject.TransformationUpdate) (*Resource, error) {
	r := l.clone()

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
		r.sd.OpenReadSeekCloser = func() (pio.ReadSeekCloser, error) {
			return u.SourceFs.Open(*u.SourceFilename)
		}
	} else if u.SourceFs != nil {
		return nil, errors.New("sourceFs is set without sourceFilename")
	}

	if u.TargetPath == "" {
		return nil, errors.New("missing targetPath")
	}

	r.paths = r.paths.FromTargetPath(u.TargetPath)
	r.mergeData(u.Data)

	return r, nil
}

func (l *Resource) mergeData(in map[string]any) {
	if len(in) == 0 {
		return
	}
	if l.sd.Data == nil {
		l.sd.Data = make(map[string]any)
	}
	for k, v := range in {
		if _, found := l.sd.Data[k]; !found {
			l.sd.Data[k] = v
		}
	}
}
