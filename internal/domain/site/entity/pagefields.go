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
	//fmt.Println("params lalala", p.RelPermalink())
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

func (p *Page) Pages() []*Page {
	ps := p.Page.Pages()
	if ps == nil {
		return make([]*Page, 0)
	}

	var pages []*Page
	for _, cp := range ps {
		np := p.clone()

		np.Page = cp
		np.PageOutput = p.getPageOutput(cp)

		pages = append(pages, np)
	}

	return pages
}

func (p *Page) getPageOutput(chp contenthub.Page) contenthub.PageOutput {
	pos, err := chp.PageOutputs()
	if err != nil {
		p.Log.Errorln("getPageOutput", err)
		panic(err)
	}

	for _, po := range pos {
		if po.TargetFormat().MediaType == p.PageOutput.TargetFormat().MediaType {
			return po
		}
	}

	p.Log.Errorln("getPageOutput", "no page output")
	panic("no page output")
}

func (p *Page) Data() map[string]any {
	return map[string]any{} //TODO for sitemap
}

func (p *Page) Content() (any, error) {
	return p.PageOutput.Content()
}

func (p *Page) IsAncestor(other any) bool {
	op, ok := other.(*Page)
	if !ok {
		return false
	}

	return p.Page.IsAncestor(op.Page)
}

func (p *Page) GitInfo() valueobject.GitInfo {
	if p.git == nil {
		return valueobject.GitInfo{}
	}

	return p.git.GetInfo(p.Page.PageFile().Filename())
}
