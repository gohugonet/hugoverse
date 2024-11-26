package entity

import (
	"fmt"
	"sort"
)

type Pages []*Page

// Len returns the number of pages in the list.
func (p Pages) Len() int {
	return len(p)
}
func (p Pages) String() string {
	return fmt.Sprintf("Pages(%d)", len(p))
}

func (p Pages) ByLastmod() Pages {
	const key = "pageSort.ByLastmod"

	date := func(p1, p2 *Page) bool {
		return p1.Lastmod().Unix() < p2.Lastmod().Unix()
	}

	pages, _ := spc.get(key, pageBy(date).Sort, p)

	return pages
}

func (p Pages) Reverse() Pages {
	const key = "pageSort.Reverse"

	reverseFunc := func(pages Pages) {
		for i, j := 0, len(pages)-1; i < j; i, j = i+1, j-1 {
			pages[i], pages[j] = pages[j], pages[i]
		}
	}

	pages, _ := spc.get(key, reverseFunc, p)

	return pages
}

type pageBy func(p1, p2 *Page) bool

func (by pageBy) Sort(pages Pages) {
	ps := &pageSorter{
		pages: pages,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Stable(ps)
}

type pageSorter struct {
	pages Pages
	by    pageBy
}

func (ps *pageSorter) Len() int      { return len(ps.pages) }
func (ps *pageSorter) Swap(i, j int) { ps.pages[i], ps.pages[j] = ps.pages[j], ps.pages[i] }

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (ps *pageSorter) Less(i, j int) bool { return ps.by(ps.pages[i], ps.pages[j]) }
