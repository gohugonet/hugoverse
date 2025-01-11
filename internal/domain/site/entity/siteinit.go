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
	menus             map[string]*lazy.Init
	taxonomies        *lazy.Init
}

func (init *siteInit) Reset() {
	init.prevNext.Reset()
	init.prevNextInSection.Reset()
	for k := range init.menus {
		init.menus[k].Reset()
	}
	init.taxonomies.Reset()
}

func (s *Site) PrepareLazyLoads() {
	initMenu := func() (any, error) {
		menus := valueobject.NewEmptyMenus()

		menusConf := s.ConfigSvc.Menus()
		for name, menu := range menusConf {
			if menus[name] == nil {
				menus[name] = valueobject.Menu{}
			}

			for _, entry := range menu {
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

		lp, err := s.GetPage(valueobject.ReservedLinksFile)

		if lp != nil && err == nil {
			hs := lp.Result().Headers()
			for _, h := range hs {
				if h.Name() == valueobject.ReservedLinksMenuSection {
					for i, l := range h.Links() {
						menus[valueobject.MenusAfter] = menus[valueobject.MenusAfter].Add(&valueobject.MenuEntry{
							MenuConfig: valueobject.MenuConfig{
								Name:   l.Text(),
								URL:    s.AbsURL(l.URL()),
								Weight: valueobject.ReservedLinksWeight + i,
							},
							Menu: valueobject.MenusAfter,
						})
					}
				}
			}
		} else if err != nil {
			s.Log.Errorf("Reserved links Menus: %v", err)
		}

		s.Navigation.menus[s.Language.currentLanguage] = menus
		return nil, nil
	}

	s.lazy = &siteInit{
		menus: map[string]*lazy.Init{},
	}

	var init lazy.Init
	for _, lang := range s.LanguageSvc.LanguageKeys() {
		s.lazy.menus[lang] = init.Branch(initMenu)
	}

	s.lazy.taxonomies = init.Branch(func() (any, error) {
		s.Navigation.taxonomies = make(TaxonomyList)

		if err := s.ContentSvc.WalkTaxonomies(s.Language.CurrentLanguageIndex(),
			func(taxonomy string, term string, page contenthub.OrdinalWeightPage) error {
				tax := s.Navigation.taxonomies[taxonomy]
				if tax == nil {
					tax = make(Taxonomy)
					s.Navigation.taxonomies[taxonomy] = tax
				}

				wp, err := s.siteWeightedPage(page)
				if err != nil {
					return err
				}

				weightedPages := tax[term]
				if weightedPages == nil {
					weightedPages = WeightedPages{wp}
					tax[term] = weightedPages
				} else {
					tax[term] = append(weightedPages, wp)
				}

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
	init, ok := s.lazy.menus[s.Language.currentLanguage]
	if ok {
		if _, err := init.Do(); err != nil {
			s.Log.Errorf("Menus: %v", err)
		}
	} else {
		s.Log.Errorf("Menus: no init for %s", s.Language.currentLanguage)
	}

	return s.Navigation.menus[s.Language.currentLanguage]
}
