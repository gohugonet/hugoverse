package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"sort"
	"strings"
)

// SortByDefault sorts pages by the default sort.
func SortByDefault(pages contenthub.Pages) {
	pageBy(lessPageTitle).Sort(pages)
}

func SortByLanguage(pages contenthub.Pages) {
	// TODO
	pageBy(lessPageTitle).Sort(pages)
}

// pageBy is a closure used in the Sort.Less method.
type pageBy func(p1, p2 contenthub.Page) bool

// Sort stable sorts the pages given the receiver's sort order.
func (by pageBy) Sort(pages contenthub.Pages) {
	ps := &pageSorter{
		pages: pages,
		by:    by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Stable(ps)
}

var (
	lessPageTitle = func(p1, p2 contenthub.Page) bool {
		return collatorStringCompare(func(p contenthub.Page) string { return p.Title() }, p1, p2) < 0
	}
)

var collatorStringCompare = func(getString func(contenthub.Page) string, p1, p2 contenthub.Page) int {
	return strings.Compare(getString(p1), getString(p2)) // TODO, use language collator
}

// A pageSorter implements the sort interface for Pages
type pageSorter struct {
	pages contenthub.Pages
	by    pageBy
}

func (ps *pageSorter) Len() int      { return len(ps.pages) }
func (ps *pageSorter) Swap(i, j int) { ps.pages[i], ps.pages[j] = ps.pages[j], ps.pages[i] }

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (ps *pageSorter) Less(i, j int) bool { return ps.by(ps.pages[i], ps.pages[j]) }
