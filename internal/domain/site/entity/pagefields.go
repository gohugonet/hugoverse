package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"time"
)

func (p *Page) Section() string {
	return p.Page.Section()
}

func (p *Page) Params() maps.Params {
	return p.Page.Params()
}

func (p *Page) Resources() PageResources {
	return p.resources
}

func (p *Page) Date() time.Time {
	return time.Now()
}

func (p *Page) PublishDate() time.Time {
	return time.Now()
}

func (p *Page) Lastmod() time.Time {
	return time.Now()
}

func (p *Page) ExpiryDate() time.Time {
	return time.Now().AddDate(1, 0, 0)
}

func (p *Page) File() contenthub.File {
	return p.Page.PageFile()
}

func (p *Page) Translations() []string {
	// TODO
	return make([]string, 0)
}

func (p *Page) OutputFormats() valueobject.OutputFormats {
	// TODO
	return make(valueobject.OutputFormats, 0)
}

func (p *Page) Sites() *sites {
	return &sites{site: p.Site}
}

type sites struct {
	site *Site
}

func (s *sites) First() *Site {
	return s.site
}

func (p *Page) Pages() contenthub.Pages {
	return nil //TODO
}
