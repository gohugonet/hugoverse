package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

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
	return &SitePager{p.page, p.Pager.Prev()}
}

func (p *SitePager) Next() *SitePager {
	return &SitePager{p.page, p.Pager.Next()}
}
