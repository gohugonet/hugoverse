package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
)

func (p *Page) Pages() contenthub.Pages {
	switch p.Kind() {
	case valueobject.KindPage:
	case valueobject.KindSection, valueobject.KindHome:
		return p.pageMap.getPagesInSection(
			pageMapQueryPagesInSection{
				pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
					Path:    p.Path(),
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
			pageMapQueryPagesInSection{
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
