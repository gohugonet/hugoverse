package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	sio "io"
	"path/filepath"
)

type PageSource struct {
	fi     fs.FileMetaInfo
	path   *paths.Path
	opener io.OpenReadSeekCloser

	contenthub.File

	stale.Staler
	cache *valueobject.Cache
}

func newPageSource(fi fs.FileMetaInfo, c *valueobject.Cache) (*PageSource, error) {
	path := paths.Parse("", fi.FileName())
	r := func() (io.ReadSeekCloser, error) {
		return fi.Open()
	}
	return &PageSource{
		fi:     fi,
		path:   path,
		opener: r,

		File: valueobject.NewFileInfo(fi),

		Staler: &stale.AtomicStaler{},
		cache:  c,
	}, nil
}

func (p *PageSource) sourceKey() string {
	return filepath.ToSlash(p.File.Filename())
}

func (p *PageSource) contentSource() ([]byte, error) {
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

func (p *PageSource) readSourceAll() ([]byte, error) {
	r, err := p.fi.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return sio.ReadAll(r)
}
