package entity

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/site"
	"github.com/mdfriday/hugoverse/internal/domain/site/valueobject"
	"github.com/mdfriday/hugoverse/pkg/compare"
	"sort"
)

type Navigation struct {
	taxonomies TaxonomyList
	menus      map[string]valueobject.Menus
}

func NewNavigation(langSvc site.LanguageService) *Navigation {
	n := &Navigation{
		taxonomies: make(TaxonomyList),
		menus:      make(map[string]valueobject.Menus),
	}
	for _, l := range langSvc.LanguageKeys() {
		n.menus[l] = valueobject.NewEmptyMenus()
	}
	return n
}

// The TaxonomyList is a list of all taxonomies and their values
// e.g. List['tags'] => TagTaxonomy (from above)
type TaxonomyList map[string]Taxonomy

// A Taxonomy is a map of keywords to a list of pages.
// For example
//
//	TagTaxonomy['technology'] = WeightedPages
//	TagTaxonomy['go']  =  WeightedPages
type Taxonomy map[string]WeightedPages

// OrderedTaxonomy is another representation of an Taxonomy using an array rather than a map.
// Important because you can't order a map.
type OrderedTaxonomy []OrderedTaxonomyEntry

// getOneOPage returns one page in the taxonomy,
// nil if there is none.
func (t OrderedTaxonomy) getOneOPage() *WeightedPage {
	if len(t) == 0 {
		return nil
	}
	return t[0].WeightedPages.Page()
}

// WeightedPages is a list of Pages with their corresponding (and relative) weight
// [{Weight: 30, Page: *1}, {Weight: 40, Page: *2}]
type WeightedPages []*WeightedPage

// Page will return the Page (of Kind taxonomyList) that represents this set
// of pages. This method will panic if p is empty, as that should never happen.
func (p WeightedPages) Page() *WeightedPage {
	if len(p) == 0 {
		_ = fmt.Errorf("page called on empty WeightedPages")
		return nil
	}

	return p[0]
}

func (p WeightedPages) Pages() []*WeightedPage {
	pages := make([]*WeightedPage, len(p))
	for i := range p {
		pages[i] = p[i]
	}
	return pages
}

func (p WeightedPages) Sort() { sort.Stable(p) }
func (p WeightedPages) Count() int {
	return len(p)
}

func (p WeightedPages) Len() int      { return len(p) }
func (p WeightedPages) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p WeightedPages) Less(i, j int) bool {
	return p[i].Weight() <= p[j].Weight()
}

// OrderedTaxonomyEntry is similar to an element of a Taxonomy, but with the key embedded (as name)
// e.g:  {Name: Technology, WeightedPages: TaxonomyPages}
type OrderedTaxonomyEntry struct {
	Name string
	WeightedPages
}

// Count returns the count the pages in this taxonomy.
func (ie OrderedTaxonomyEntry) Count() int {
	return ie.WeightedPages.Count()
}

// Term returns the name given to this taxonomy.
func (ie OrderedTaxonomyEntry) Term() string {
	return ie.Name
}

// ByCount returns an ordered taxonomy sorted by # of pages per key.
// If taxonomies have the same # of pages, sort them alphabetical
func (i Taxonomy) ByCount() OrderedTaxonomy {
	count := func(i1, i2 *OrderedTaxonomyEntry) bool {
		li1 := len(i1.WeightedPages)
		li2 := len(i2.WeightedPages)

		if li1 == li2 {
			return compare.LessStrings(i1.Name, i2.Name)
		}
		return li1 > li2
	}

	ia := i.TaxonomyArray()
	oiBy(count).Sort(ia)
	return ia
}

// TaxonomyArray returns an ordered taxonomy with a non defined order.
func (i Taxonomy) TaxonomyArray() OrderedTaxonomy {
	ies := make([]OrderedTaxonomyEntry, len(i))
	count := 0
	for k, v := range i {
		ies[count] = OrderedTaxonomyEntry{Name: k, WeightedPages: v}
		count++
	}
	return ies
}

// Closure used in the Sort.Less method.
type oiBy func(i1, i2 *OrderedTaxonomyEntry) bool

func (by oiBy) Sort(taxonomy OrderedTaxonomy) {
	ps := &orderedTaxonomySorter{
		taxonomy: taxonomy,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Stable(ps)
}

// A type to implement the sort interface for TaxonomyEntries.
type orderedTaxonomySorter struct {
	taxonomy OrderedTaxonomy
	by       oiBy
}

// Len is part of sort.Interface.
func (s *orderedTaxonomySorter) Len() int {
	return len(s.taxonomy)
}

// Swap is part of sort.Interface.
func (s *orderedTaxonomySorter) Swap(i, j int) {
	s.taxonomy[i], s.taxonomy[j] = s.taxonomy[j], s.taxonomy[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *orderedTaxonomySorter) Less(i, j int) bool {
	return s.by(&s.taxonomy[i], &s.taxonomy[j])
}
