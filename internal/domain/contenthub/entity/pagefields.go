package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"strings"
)

func (p *Page) Pages() contenthub.Pages {
	switch p.Kind() {
	case valueobject.KindPage:
	case valueobject.KindSection, valueobject.KindHome:
		return p.pageMap.getPagesInSection(
			pageMapQueryPagesInSection{
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

func (p *Page) IsAncestor(other contenthub.Page) bool {
	if other.Path() == p.Path() {
		return false
	}

	return strings.HasPrefix(other.Path(), paths.AddTrailingSlash(p.Path()))
}
