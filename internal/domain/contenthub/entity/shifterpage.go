package entity

import (
	"github.com/gohugonet/hugoverse/pkg/doctree"
)

type PageShifter struct {
	*Shifter
}

func (s *PageShifter) Shift(n *PageTreesNode, dimension doctree.Dimension, exact bool) (*PageTreesNode, bool, doctree.DimensionFlag) {
	newNode, found := n.shift(dimension[doctree.DimensionLanguage.Index()], exact)
	if newNode != nil {
		if found {
			return newNode, true, doctree.DimensionLanguage
		}
	}

	return nil, false, doctree.DimensionNone
}
