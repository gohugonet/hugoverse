package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
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

	s.lazy.menus = init.Branch(func() (any, error) {
		menus := valueobject.Menus{}

		menusConf := s.ConfigSvc.Menus()

		for name, menu := range menusConf {
			menus[name] = valueobject.Menu{}

			for _, entry := range menu {
				if menus[name] == nil {
					menus[name] = valueobject.Menu{}
				}

				menus[name] = menus[name].Add(&valueobject.MenuEntry{
					MenuConfig: valueobject.MenuConfig{
						Name:   entry.Name(),
						URL:    entry.URL(),
						Weight: entry.Weight(),
					},
					Menu: name,
				})
			}
		}
		s.Navigation.menus = menus

		return nil, nil
	})

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
		s.Log.Errorf("Taxonomies: %v", err)
	}
	return s.Navigation.taxonomies
}

func (s *Site) Menus() valueobject.Menus {
	if _, err := s.lazy.menus.Do(); err != nil {
		s.Log.Errorf("Menus: %v", err)
	}

	return s.Navigation.menus
}
