package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/lazy"
)

// Lazily loaded site dependencies.
type siteInit struct {
	prevNext          *lazy.Init
	prevNextInSection *lazy.Init
	menus             *lazy.Init
	taxonomies        *lazy.Init
}

func (init *siteInit) Reset() {
	init.prevNext.Reset()
	init.prevNextInSection.Reset()
	init.menus.Reset()
	init.taxonomies.Reset()
}

func (s *Site) PrepareLazyLoads() {
	s.lazy = &siteInit{}

	var init lazy.Init

	s.lazy.taxonomies = init.Branch(func() (any, error) {
		s.Navigation.taxonomies = make(TaxonomyList)

		if err := s.ContentSvc.WalkTaxonomies(s.Language.CurrentLanguageIndex(),
			func(taxonomy string, term string, page contenthub.OrdinalWeightPage) error {
				tax := s.Navigation.taxonomies[taxonomy]
				if tax == nil {
					tax = make(Taxonomy)
					s.Navigation.taxonomies[taxonomy] = tax
				}
				tax[term] = page

				return nil
			}); err != nil {
			return nil, err
		}
		return s.Navigation.taxonomies, nil
	})
}

func (s *Site) Taxonomies() TaxonomyList {
	if _, err := s.lazy.taxonomies.Do(); err != nil {
		panic(err)
	}
	return s.Navigation.taxonomies
}
