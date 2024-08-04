package entity

import (
	"github.com/gohugonet/hugoverse/pkg/doctree"
)

type SourceShifter struct{}

func (s *SourceShifter) Delete(n *PageTreesNode, dimension doctree.Dimension) (bool, bool) {
	wasDeleted := n.delete(dimension[doctree.DimensionLanguage.Index()])
	return wasDeleted, n.isEmpty()
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

func (s *SourceShifter) ForEachInDimension(n *PageTreesNode, d int, f func(*PageTreesNode) bool) {
	if d != doctree.DimensionLanguage.Index() {
		panic("only language dimension supported")
	}
	f(n)
}

func (s *SourceShifter) InsertInto(old, new *PageTreesNode, dimension doctree.Dimension) *PageTreesNode {
	return old.mergeWithLang(new, dimension[doctree.DimensionLanguage.Index()])
}

func (s *SourceShifter) Insert(old, new *PageTreesNode) *PageTreesNode {
	return old.merge(new)
}
