package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	sio "io"
	"path/filepath"
	"sync/atomic"
)

var pageIDCounter atomic.Uint64

type Source struct {
	*valueobject.Identity
	*valueobject.File

	// Returns the position in bytes after any front matter.
	posMainContent int

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

func (p *Source) registerHandler(fm valueobject.ItemSourceHandler,
	summary valueobject.IterHandler, bytes valueobject.ItemHandler,
	shortcode valueobject.IterHandler) {

	p.parseInfo.FrontMatterHandler = fm
	p.parseInfo.SummaryHandler = summary
	p.parseInfo.BytesHandler = bytes
	p.parseInfo.ShortcodeHandler = shortcode
}

func (p *Source) parse() error {
	content, err := p.contentSource()
	if err != nil {
		return err
	}

	items, err := pageparser.ParseBytes(
		content,
		pageparser.Config{},
	)
	if err != nil {
		return err
	}

	p.parseInfo.ItemsStep1 = items

	if err := p.parseInfo.Handle(); err != nil {
		return err
	}

	return nil
}
