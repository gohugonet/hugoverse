package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"sync"
)

type PaginatorSvc interface {
	PageSize() int
	GlobalRegularPages() contenthub.Pages
}

type Paginator struct {
	current contenthub.Pager

	init sync.Once
	svc  PaginatorSvc
	page contenthub.Page
}

func NewPaginator(svc PaginatorSvc, page contenthub.Page) *Paginator {
	return &Paginator{
		svc:  svc,
		page: page,
	}
}

func (p *Paginator) Current() contenthub.Pager {
	return p.current
}

func (p *Paginator) SetCurrent(current contenthub.Pager) {
	p.current = current
}

func (p *Paginator) Paginate(groups contenthub.PageGroups) (contenthub.Pager, error) {
	var initErr error
	p.init.Do(func() {
		pagerSize := p.svc.PageSize()

		paginator, err := valueobject.NewPaginatorFromPageGroups(groups, pagerSize, p.page.Paths().Base())
		if err != nil {
			initErr = err
			return
		}

		p.current = paginator.Pagers()[0]
	})

	if initErr != nil {
		return nil, initErr
	}

	return p.current, nil
}

func (p *Paginator) Paginator() (contenthub.Pager, error) {
	var initErr error
	p.init.Do(func() {
		pagerSize := p.svc.PageSize()

		var pages contenthub.Pages

		switch p.page.Kind() {
		case valueobject.KindHome:
			pages = p.svc.GlobalRegularPages()
		case valueobject.KindTerm, valueobject.KindTaxonomy:
			pages = p.page.Pages(p.page.PageIdentity().PageLanguageIndex())
		default:
			pages = p.page.RegularPages()
			if pages == nil {
				pages = p.svc.GlobalRegularPages()
			}
		}

		paginator, err := valueobject.NewPaginatorFromPages(pages, pagerSize, p.page.Paths().Base())
		if err != nil {
			initErr = err
			return
		}

		p.current = paginator.Pagers()[0]
	})

	if initErr != nil {
		return nil, initErr
	}

	return p.current, nil
}
