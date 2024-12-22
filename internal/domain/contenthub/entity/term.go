package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/types"
	"github.com/spf13/cast"
)

type Term struct {
	Terms map[string][]string

	FsSvc contenthub.FsService
	Cache *Cache
}

func (t *Term) Assemble(pages *doctree.NodeShiftTree[*PageTreesNode],
	entries *doctree.TreeShiftTree[*WeightedTermTreeNode],
	pb *PageBuilder) error {

	lockType := doctree.LockTypeWrite
	w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree:     pages,
		LockType: lockType,
		Handle: func(s string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {

			ps, found := n.getPage()
			if !found {
				return false, nil
			}

			for _, viewName := range pb.Taxonomy.Views {
				vals := types.ToStringSlicePreserveString(valueobject.GetParam(ps.Params(), viewName.Plural(), false))
				if vals == nil {
					continue
				}

				w := valueobject.GetParamToLower(ps.Params(), viewName.Plural()+"_weight")
				weight, err := cast.ToIntE(w)
				if err != nil {
					pb.Log.Warnf("Unable to convert taxonomy weight %#v to int for %q", w, ps.Paths().Path())
					// weight will equal zero, so let the flow continue
				}

				for i, v := range vals {
					if v == "" {
						continue
					}

					viewTermKey := "/" + viewName.Plural() + "/" + v

					fmi := t.FsSvc.NewFileMetaInfo(viewTermKey + "/_index.md")
					f, err := valueobject.NewFileInfo(fmi)
					if err != nil {
						return false, err
					}
					pi := f.Paths()

					term := pages.Get(pi.Base())

					if term == nil {
						ps, err := newPageSource(f, t.Cache)
						if err != nil {
							return false, err
						}

						p, err := pb.WithSource(ps).KindBuild()
						if err != nil {
							return false, err
						}

						pages.InsertIntoValuesDimension(pi.Base(), newPageTreesNode(p))
						term = pages.Get(pi.Base())
					} else {
						tp, found := term.getPage()
						if !found {
							return false, nil
						}

						m := tp.(*TermPage)
						m.term = v
						m.singular = viewName.Singular()
					}

					if s == "" {
						// Consider making this the real value.
						s = "/"
					}

					key := pi.Base() + s

					tp, found := term.getPage()
					if !found {
						return false, nil
					}

					entries.Insert(key, &WeightedTermTreeNode{
						PageTreesNode: newPageTreesNode(ps),
						term:          &ordinalWeightPage{Page: tp.(*TermPage), ordinal: i, weight: weight},
					})
				}
			}
			return false, nil
		},
	}

	return w.Walk(context.Background())
}
