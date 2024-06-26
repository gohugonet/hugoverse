package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/paths"
	sio "io"
	"path/filepath"
)

type Source struct {
	fi     fs.FileMetaInfo
	path   *paths.Path
	opener io.OpenReadSeekCloser

	contenthub.File

	// Returns the position in bytes after any front matter.
	posMainContent int
	parseInfo      *valueobject.SourceParseInfo

	stale.Staler
	cache *valueobject.Cache
}

func newPageSource(fi fs.FileMetaInfo, c *valueobject.Cache) (*Source, error) {
	path := paths.Parse("", fi.FileName())
	r := func() (io.ReadSeekCloser, error) {
		return fi.Open()
	}
	return &Source{
		fi:     fi,
		path:   path,
		opener: r,

		File: valueobject.NewFileInfo(fi),

		Staler: &stale.AtomicStaler{},
		cache:  c,

		parseInfo: &valueobject.SourceParseInfo{},
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
	r, err := p.fi.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return sio.ReadAll(r)
}

func (p *Source) registerHandler(fm valueobject.ItemSourceHandler,
	summary valueobject.IterHandler, bytes valueobject.ItemHandler) {

	p.parseInfo.FrontMatterHandler = fm
	p.parseInfo.SummaryHandler = summary
	p.parseInfo.BytesHandler = bytes
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
