package entity

import (
	"context"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/doctree"
	"github.com/mdfriday/hugoverse/pkg/paths"
	"strings"
)

func (ch *ContentHub) WalkPages(langIndex int, walker contenthub.WalkFunc) error {
	tree := ch.PageMap.TreePages.Shape(0, langIndex)

	w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree: tree,
		Handle: func(key string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {

			ps, found := n.getPage()
			if !found {
				return false, nil
			}

			if err := walker(ps); err != nil {
				return false, err
			}

			return false, nil
		},
	}

	if err := w.Walk(context.Background()); err != nil {
		return err
	}

	return nil
}

func (ch *ContentHub) WalkTaxonomies(langIndex int, walker contenthub.WalkTaxonomyFunc) error {
	tree := ch.PageMap.TreePages.Shape(0, langIndex)

	tc := ch.PageMap.PageBuilder.Taxonomy
	for _, viewName := range tc.Views {
		key := tc.PluralTreeKey(viewName.Plural())

		w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
			Tree:     tree,
			Prefix:   paths.AddTrailingSlash(key),
			LockType: doctree.LockTypeRead,
			Handle: func(s string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
				p, found := n.getPage()
				if !found {
					return false, nil
				}

				switch p.Kind() {
				case valueobject.KindTerm:
					t := p.(*TermPage)

					if t.term == "" {
						panic("term is empty")
					}
					k := strings.ToLower(t.term)

					err := ch.PageMap.TreeTaxonomyEntries.WalkPrefix(
						doctree.LockTypeRead,
						paths.AddTrailingSlash(s),
						func(ss string, wn *WeightedTermTreeNode) (bool, error) {
							sp, found := wn.getPage()
							if !found {
								return false, nil
							}

							if err := walker(viewName.Plural(), k,
								&weightPage{
									ordinalWeightPage: wn.term,
									page:              sp},
							); err != nil {
								return false, err
							}

							return false, nil
						},
					)
					if err != nil {
						return true, err
					}

				default:
					return false, nil
				}

				return false, nil
			},
		}

		if err := w.Walk(context.Background()); err != nil {
			return err
		}
	}

	return nil
}
