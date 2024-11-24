package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"strings"
)

func (p *Page) Pages(langIndex int) contenthub.Pages {
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
				Path: p.Path(),
			},
		)
	case valueobject.KindTaxonomy:
		return p.pageMap.getPagesInSection(
			langIndex,
			pageMapQueryPagesInSection{
				Index: langIndex,
				pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
					Path:    p.Path(),
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

func (p *Page) Terms(langIndex int, taxonomy string) contenthub.Pages {
	return p.pageMap.getTermsForPageInTaxonomy(p.Paths().Base(), taxonomy)
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
