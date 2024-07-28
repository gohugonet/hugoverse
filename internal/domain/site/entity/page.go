package entity

import "github.com/gohugonet/hugoverse/internal/domain/contenthub"

type Pages []Page

// Len returns the number of pages in the list.
func (p Pages) Len() int {
	return len(p)
}

type Page struct {
	contenthub.Page
}
