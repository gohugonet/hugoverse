package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
)

// The TaxonomyList is a list of all taxonomies and their values
// e.g. List['tags'] => TagTaxonomy (from above)
type TaxonomyList map[string]Taxonomy

// A Taxonomy is a map of keywords to a list of pages.
// For example
//
//	TagTaxonomy['technology'] = WeightedPages
//	TagTaxonomy['go']  =  WeightedPages
type Taxonomy map[string]contenthub.OrdinalWeightPage

type Navigation struct {
	taxonomies TaxonomyList
	menus      valueobject.Menus
}
