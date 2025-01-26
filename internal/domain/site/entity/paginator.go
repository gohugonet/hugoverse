package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

func (p *Page) PrevInSection() *Page {
	cp := p.Page.PrevInSection()
	if cp == nil {
		return nil
	}
	sp, err := p.sitePage(cp)
	if err != nil {
		p.Log.Errorf("NextInSection for page %s: %v", p.Path(), err)
		return nil
	}
	return sp
}

func (p *Page) NextInSection() *Page {
	cp := p.Page.NextInSection()
	if cp == nil {
		return nil
	}
	sp, err := p.sitePage(cp)
	if err != nil {
		p.Log.Errorf("NextInSection for page %s: %v", p.Path(), err)
		return nil
	}
	return sp
}

func (p *Page) Paginator() (*SitePager, error) {
	pager, err := p.Page.Paginator()
	if err != nil {
		return nil, err
	}

	return &SitePager{p, pager}, nil
}

func (p *Page) Paginate(groups contenthub.PageGroups) (*SitePager, error) {
	pager, err := p.Page.Paginate(groups)
	if err != nil {
		return nil, err
	}

	return &SitePager{p, pager}, nil
}

type SitePager struct {
	page *Page
	contenthub.Pager
}

func (p *SitePager) Pages() Pages {
	return p.page.sitePages(p.Pager.Pages())
}

func (p *SitePager) Prev() *SitePager {
	if !p.Pager.HasPrev() {
		return nil
	}
	return &SitePager{p.page, p.Pager.Prev()}
}

func (p *SitePager) Next() *SitePager {
	if !p.Pager.HasNext() {
		return nil
	}
	return &SitePager{p.page, p.Pager.Next()}
}
