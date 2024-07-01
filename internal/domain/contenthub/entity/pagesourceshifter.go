package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/doctree"
)

type SourceShifter struct {
}

func (s *SourceShifter) Delete(n *PageTreesNode, dimension doctree.Dimension) (bool, bool) {
	lidx := dimension[0]
	switch v := n.(type) {
	case contentNodeIs:
		stale.MarkStale(v[lidx])
		wasDeleted := v[lidx] != nil
		v[lidx] = nil
		isEmpty := true
		for _, vv := range v {
			if vv != nil {
				isEmpty = false
				break
			}
		}
		return wasDeleted, isEmpty
	case resourceSources:
		stale.MarkStale(v[lidx])
		wasDeleted := v[lidx] != nil
		v[lidx] = nil
		isEmpty := true
		for _, vv := range v {
			if vv != nil {
				isEmpty = false
				break
			}
		}
		return wasDeleted, isEmpty
	case *resourceSource:
		if lidx != v.LangIndex() {
			return false, false
		}
		resource.MarkStale(v)
		return true, true
	case *pageState:
		if lidx != v.s.languagei {
			return false, false
		}
		resource.MarkStale(v)
		return true, true
	default:
		panic(fmt.Sprintf("unknown type %T", n))
	}
}

func (s *SourceShifter) Shift(n *PageTreesNode, dimension doctree.Dimension, exact bool) (*PageTreesNode, bool, doctree.DimensionFlag) {
	newNode := n.shift(dimension[doctree.DimensionLanguage.Index()], exact)
	if newNode != nil {
		return newNode, true, doctree.DimensionLanguage
	}

	return nil, false, 0
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
