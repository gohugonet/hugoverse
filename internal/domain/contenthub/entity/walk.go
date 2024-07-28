package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/doctree"
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
