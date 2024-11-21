package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/compare"
	"sort"
)

type Navigation struct {
	taxonomies TaxonomyList
	menus      valueobject.Menus
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
func (t OrderedTaxonomy) getOneOPage() contenthub.Page {
	if len(t) == 0 {
		return nil
	}
	return t[0].WeightedPages.Page()
}

// WeightedPages is a list of Pages with their corresponding (and relative) weight
// [{Weight: 30, Page: *1}, {Weight: 40, Page: *2}]
type WeightedPages []contenthub.OrdinalWeightPage

// Page will return the Page (of Kind taxonomyList) that represents this set
// of pages. This method will panic if p is empty, as that should never happen.
func (p WeightedPages) Page() contenthub.Page {
	if len(p) == 0 {
		panic("WeightedPages is empty")
	}

	return p[0]
}

// OrderedTaxonomyEntry is similar to an element of a Taxonomy, but with the key embedded (as name)
// e.g:  {Name: Technology, WeightedPages: TaxonomyPages}
type OrderedTaxonomyEntry struct {
	Name string
	WeightedPages
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
