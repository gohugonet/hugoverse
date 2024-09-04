package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/identity"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

var (
	_ resources.Resource               = (*Resource)(nil)
	_ resources.ReadSeekCloserResource = (*Resource)(nil)
	_ hashProvider                     = (*Resource)(nil)
)

type Resource struct {
	stale.Staler

	h *valueobject.ResourceHash // A hash of the source content. Is only calculated in caching situations.

	openReadSeekCloser pio.OpenReadSeekCloser
	mediaType          media.Type

	paths valueobject.ResourcePaths

	data              map[string]any
	dependencyManager identity.Manager
}

func (l *Resource) Name() string {
	return l.paths.File
}

func (l *Resource) NameNormalized() string {
	return paths.ToSlashPreserveLeading(l.paths.TargetPath())
}

func (l *Resource) MediaType() media.Type {
	return l.mediaType
}

func (l *Resource) ResourceType() string {
	return l.MediaType().MainType
}

func (l *Resource) RelPermalink() string {
	// TODO: use config BaseURL
	return paths.PathEscape(l.paths.TargetLink())
}

func (l *Resource) Permalink() string {
	// TODO: use config BaseURL
	return paths.PathEscape(l.paths.TargetPath())
}
func (l *Resource) TargetPath() string {
	return l.paths.TargetPath()
}

func (l *Resource) Data() any {
	return l.data
}

func (l *Resource) Err() resources.ResourceError {
	return nil
}

func (l *Resource) ReadSeekCloser() (pio.ReadSeekCloser, error) {
	return l.openReadSeekCloser()
}

func (l *Resource) Content(context.Context) (any, error) {
	r, err := l.ReadSeekCloser()
	if err != nil {
		return "", err
	}
	defer r.Close()

	return pio.ReadString(r)
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
	// TODO, use config BaseURL
	return l.RelPermalink()
}

func (l *Resource) DependencyManager() identity.Manager {
	return l.dependencyManager
}

func (l *Resource) meta() valueobject.ResourceMetadata {
	return valueobject.ResourceMetadata{
		MediaTypeV: l.mediaType.Type,
		Target:     l.paths.TargetPath(),
		MetaData:   l.data,
	}
}

func (l *Resource) mergeData(in map[string]any) {
	if len(in) == 0 {
		return
	}
	if l.data == nil {
		l.data = make(map[string]any)
	}
	for k, v := range in {
		if _, found := l.data[k]; !found {
			l.data[k] = v
		}
	}
}
