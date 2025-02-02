package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"strings"
)

func (p *Page) Pages() contenthub.Pages {
	langIndex := p.PageIdentity().PageLanguageIndex()

	switch p.Kind() {
	case valueobject.KindPage:
	case valueobject.KindSection, valueobject.KindHome:
		return p.pageMap.getPagesInSection(
			langIndex,
			pageMapQueryPagesInSection{
				Index: langIndex,
				pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
					Path:    p.Paths().Base(),
					KeyPart: "page-section",
					Include: pagePredicates.ShouldListLocal.And(
						pagePredicates.KindPage.Or(pagePredicates.KindSection),
					),
				},
			},
		)
	case valueobject.KindTerm:
		return p.pageMap.getPagesWithTerm(
			pageMapQueryPagesBelowPath{
				Path: p.Paths().Base(),
			},
		)
	case valueobject.KindTaxonomy:
		return p.pageMap.getPagesInSection(
			langIndex,
			pageMapQueryPagesInSection{
				Index: langIndex,
				pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
					Path:    p.Paths().Base(),
					KeyPart: "term",
					Include: pagePredicates.ShouldListLocal.And(pagePredicates.KindTerm),
				},
				Recursive: true,
			},
		)
	default:
		return nil
	}

	return nil
}

func (p *Page) RegularPages() contenthub.Pages {
	switch p.Kind() {
	case valueobject.KindPage:
	case valueobject.KindSection, valueobject.KindHome, valueobject.KindTaxonomy:
		return p.pageMap.getPagesInSection(
			p.PageIdentity().PageLanguageIndex(),
			pageMapQueryPagesInSection{
				Index: p.PageIdentity().PageLanguageIndex(),
				pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
					Path:    p.Paths().Base(),
					Include: pagePredicates.ShouldListLocal.And(pagePredicates.KindPage),
				},
			},
		)
	case valueobject.KindTerm:
		return p.pageMap.getPagesWithTerm(
			pageMapQueryPagesBelowPath{
				Path:    p.Paths().Base(),
				Include: pagePredicates.ShouldListLocal.And(pagePredicates.KindPage),
			},
		)
	default:
		return nil
	}
	return nil
}

func (p *Page) Sections(langIndex int) contenthub.Pages {
	prefix := paths.AddTrailingSlash(p.Paths().Base())
	return p.pageMap.getSections(langIndex, prefix)
}

func (p *Page) Terms(langIndex int, taxonomy string) contenthub.Pages {
	return p.pageMap.getTermsForPageInTaxonomy(p.Paths().Base(), taxonomy)
}

func (p *Page) IsTranslated() bool {
	return len(p.Translations()) > 0
}

func (p *Page) Translations() contenthub.Pages {
	key := p.Path() + "/" + p.PageLanguage() + "/" + "translations"
	pages, err := p.pageMap.getOrCreatePagesFromCache(nil, key, func(string) (contenthub.Pages, error) {
		var pas contenthub.Pages
		for _, pp := range p.AllTranslations() {
			if !pp.Eq(p) {
				pas = append(pas, pp)
			}
		}
		return pas, nil
	})
	if err != nil {
		panic(err)
	}
	return pages
}

// AllTranslations returns all translations, including the current Page.
func (p *Page) AllTranslations() contenthub.Pages {
	key := p.Path() + "/" + "translations-all"
	// This is called from Translations, so we need to use a different partition, cachePages2,
	// to avoid potential deadlocks.
	pages, err := p.pageMap.getOrCreatePagesFromCache(p.pageMap.Cache.CachePages2, key, func(string) (contenthub.Pages, error) {
		var pas contenthub.Pages
		p.pageMap.TreePages.ForEachInDimension(p.Paths().Base(), doctree.DimensionLanguage.Index(),
			func(n *PageTreesNode) bool {
				if n != nil {
					pas = n.getPages()
				}
				return false
			},
		)

		pas = pagePredicates.ShouldLink.Filter(pas)
		valueobject.SortByLanguage(pas)
		return pas, nil
	})
	if err != nil {
		panic(err)
	}

	return pages
}

func (p *Page) IsAncestor(other contenthub.Page) bool {
	if other.Path() == p.Path() {
		return false
	}

	return strings.HasPrefix(other.Path(), paths.AddTrailingSlash(p.Path()))
}

func (p *Page) Title() string {
	return p.title
}

func (p *Page) Parent() contenthub.Page {
	if p.IsHome() {
		return nil
	}

	dir := p.Paths().ContainerDir()

	if dir == "" {
		return nil
	}

	for {
		_, n := p.pageMap.TreePages.LongestPrefix(dir, true, nil)
		if n == nil {
			return nil
		}

		pg, found := n.getPage()
		if p.IsBundled() {
			if found {
				return pg
			}

			return nil
		}

		if !pg.IsPage() {
			if found {
				return pg
			}
			return nil
		}

		dir = paths.Dir(dir)
	}
}

func (p *Page) PrevInSection() contenthub.Page {
	langIndex := p.PageIdentity().PageLanguageIndex()
	ps := p.pageMap.getPagesInSection(
		langIndex,
		pageMapQueryPagesInSection{
			Index: langIndex,
			pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
				Path:    p.Paths().ContainerDir(),
				KeyPart: "page-section",
				Include: pagePredicates.ShouldListLocal.And(
					pagePredicates.KindPage.Or(pagePredicates.KindSection),
				),
			},
		},
	)

	if len(ps) == 0 {
		return nil
	}

	currentPageId := p.PageIdentity().IdentifierBase()
	for i := 0; i < len(ps); i++ {
		if ps[i].PageIdentity().IdentifierBase() == currentPageId {
			if i > 0 {
				return ps[i-1]
			}
			return nil
		}
	}

	return nil
}

func (p *Page) NextInSection() contenthub.Page {
	langIndex := p.PageIdentity().PageLanguageIndex()
	ps := p.pageMap.getPagesInSection(
		langIndex,
		pageMapQueryPagesInSection{
			Index: langIndex,
			pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
				Path:    p.Paths().ContainerDir(),
				KeyPart: "page-section",
				Include: pagePredicates.ShouldListLocal.And(
					pagePredicates.KindPage.Or(pagePredicates.KindSection),
				),
			},
		},
	)

	if len(ps) == 0 {
		return nil
	}

	currentPageId := p.PageIdentity().IdentifierBase()
	for i := 0; i < len(ps); i++ {
		if ps[i].PageIdentity().IdentifierBase() == currentPageId {
			if i < len(ps)-1 {
				return ps[i+1]
			}
			return nil
		}
	}

	return nil
}

func (p *Page) Store() *maps.Scratch {
	if p.store == nil {
		p.store = maps.NewScratch()
	}
	return p.store
}

func (p *Page) Scratch() *maps.Scratch {
	return p.Store()
}
