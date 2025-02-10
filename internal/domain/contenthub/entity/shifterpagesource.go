package entity

import (
	"github.com/mdfriday/hugoverse/pkg/doctree"
)

type SourceShifter struct {
	*Shifter
}

func (s *SourceShifter) Shift(n *PageTreesNode, dimension doctree.Dimension, exact bool) (*PageTreesNode, bool, doctree.DimensionFlag) {
	newNode, found := n.shift(dimension[doctree.DimensionLanguage.Index()], exact)
	if newNode != nil {
		if found {
			return newNode, true, doctree.DimensionLanguage
		}
		return newNode, true, doctree.DimensionNone
	}

	return nil, false, doctree.DimensionNone
}
