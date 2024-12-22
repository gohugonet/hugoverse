package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"path"
)

type pagers []*Pager

type Pager struct {
	number int
	*Paginator
}

func (p *Pager) element() paginatedElement {
	if len(p.paginatedElements) == 0 {
		return paginatorEmptyPages
	}
	return p.paginatedElements[p.PageNumber()-1]
}

func (p *Pager) NumberOfElements() int {
	return p.element().Len()
}

func (p *Pager) PageNumber() int {
	return p.number
}

func (p *Pager) Pages() contenthub.Pages {
	if len(p.paginatedElements) == 0 {
		return paginatorEmptyPages
	}

	if pages, ok := p.element().(contenthub.Pages); ok {
		return pages
	}

	return paginatorEmptyPages
}

// HasPrev tests whether there are page(s) before the current.
func (p *Pager) HasPrev() bool {
	return p.PageNumber() > 1
}

// Prev returns the pager for the previous page.
func (p *Pager) Prev() contenthub.Pager {
	if !p.HasPrev() {
		return nil
	}
	return p.pagers[p.PageNumber()-2]
}

// HasNext tests whether there are page(s) after the current.
func (p *Pager) HasNext() bool {
	return p.PageNumber() < len(p.paginatedElements)
}

// Next returns the pager for the next page.
func (p *Pager) Next() contenthub.Pager {
	if !p.HasNext() {
		return nil
	}
	return p.pagers[p.PageNumber()]
}

func (p *Pager) URL() string {
	pageNumber := p.PageNumber()
	if pageNumber > 1 {
		rel := fmt.Sprintf("/%s/%d/", "page", pageNumber)
		return path.Join(p.base, rel)
	}

	return p.base
}
