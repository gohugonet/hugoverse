package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
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
