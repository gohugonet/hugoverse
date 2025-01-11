package valueobject

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"math"
)

type paginatedElement interface {
	Len() int
}

var (
	paginatorEmptyPages contenthub.Pages
)

type Paginator struct {
	paginatedElements []paginatedElement
	pagers

	base  string
	total int
	size  int
}

func NewPaginatorFromPageGroups(pageGroups contenthub.PageGroups, size int, base string) (*Paginator, error) {
	if size <= 0 {
		return nil, errors.New("paginator size must be positive")
	}

	split := splitPageGroups(pageGroups, size)

	return newPaginator(split, pageGroups.Len(), size, base)
}

func NewPaginatorFromPages(pages contenthub.Pages, size int, base string) (*Paginator, error) {
	if size <= 0 {
		return nil, errors.New("paginator size must be positive")
	}

	split := splitPages(pages, size)

	return newPaginator(split, len(pages), size, base)
}

func newPaginator(elements []paginatedElement, total, size int, base string) (*Paginator, error) {
	p := &Paginator{total: total, paginatedElements: elements, size: size, base: base}

	var ps pagers

	if len(elements) > 0 {
		ps = make(pagers, len(elements))
		for i := range p.paginatedElements {
			ps[i] = &Pager{number: i + 1, Paginator: p}
		}
	} else {
		ps = make(pagers, 1)
		ps[0] = &Pager{number: 1, Paginator: p}
	}

	p.pagers = ps

	return p, nil
}

func splitPages(pages contenthub.Pages, size int) []paginatedElement {
	var split []paginatedElement
	for low, j := 0, len(pages); low < j; low += size {
		high := int(math.Min(float64(low+size), float64(len(pages))))
		split = append(split, pages[low:high])
	}

	return split
}

func splitPageGroups(pageGroups contenthub.PageGroups, size int) []paginatedElement {
	type keyPage struct {
		key  string
		page contenthub.Page
	}

	var (
		split     []paginatedElement
		flattened []keyPage
	)

	for _, g := range pageGroups {
		for _, p := range g.Pages() {
			flattened = append(flattened, keyPage{g.Key(), p})
		}
	}

	numPages := len(flattened)

	for low, j := 0, numPages; low < j; low += size {
		high := int(math.Min(float64(low+size), float64(numPages)))

		var (
			pg         contenthub.PageGroups
			key        string
			groupIndex = -1
		)

		for k := low; k < high; k++ {
			kp := flattened[k]
			if key == "" || key != kp.key {
				key = kp.key
				pg = append(pg, &PageGroup{key: key})
				groupIndex++
			}
			pg[groupIndex].Append(kp.page)
		}
		split = append(split, pg)
	}

	return split
}

func (p *Paginator) Pagers() contenthub.Pagers {
	var ps contenthub.Pagers

	for _, pager := range p.pagers {
		ps = append(ps, pager)
	}

	return ps
}

func (p *Paginator) TotalPages() int {
	return len(p.paginatedElements)
}
