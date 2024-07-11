package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	sio "io"
	"path/filepath"
	"sync/atomic"
)

var pageIDCounter atomic.Uint64

type Source struct {
	*valueobject.Identity
	*valueobject.File

	stale.Staler
	cache *valueobject.Cache
}

func newPageSource(fi *valueobject.File, c *valueobject.Cache) (*Source, error) {
	return &Source{
		Identity: &valueobject.Identity{
			Id: pageIDCounter.Add(1),
		},

		File: fi,

		Staler: &stale.AtomicStaler{},
		cache:  c,
	}, nil
}

func (p *Source) sourceKey() string {
	return filepath.ToSlash(p.File.Filename())
}

func (p *Source) contentSource() ([]byte, error) {
	key := p.sourceKey()
	v, err := p.cache.CacheContentSource.GetOrCreate(key, func(string) (*stale.Value[[]byte], error) {
		b, err := p.readSourceAll()
		if err != nil {
			return nil, err
		}

		return &stale.Value[[]byte]{
			Value: b,
			IsStaleFunc: func() bool {
				return p.Staler.IsStale()
			},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return v.Value, nil
}

func (p *Source) readSourceAll() ([]byte, error) {
	r, err := p.File.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return sio.ReadAll(r)
}
