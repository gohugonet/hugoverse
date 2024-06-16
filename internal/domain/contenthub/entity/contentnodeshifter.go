package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/identity"
)

var _ contenthub.ContentNode = (*contentNodeIs)(nil)

type contentNodeIs []contenthub.ContentNode

func (n contentNodeIs) Path() string {
	return n[0].Path()
}

func (n contentNodeIs) isContentNodeBranch() bool {
	return n[0].IsContentNodeBranch()
}

func (n contentNodeIs) GetIdentity() identity.Identity {
	return n[0].GetIdentity()
}

func (n contentNodeIs) ForEeachIdentity(f func(identity.Identity) bool) bool {
	for _, nn := range n {
		if nn != nil {
			if nn.ForEeachIdentity(f) {
				return true
			}
		}
	}
	return false
}

func (n contentNodeIs) resetBuildState() {
	for _, nn := range n {
		if nn != nil {
			nn.ResetBuildState()
		}
	}
}

func (n contentNodeIs) MarkStale() {
	for _, nn := range n {
		stale.MarkStale(nn)
	}
}

type ContentNodeShifter struct {
	NumLanguages int
}

func (s *ContentNodeShifter) Delete(n contentNodeI, dimension doctree.Dimension) (bool, bool) {
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

func (s *ContentNodeShifter) Shift(n contentNodeI, dimension doctree.Dimension, exact bool) (contentNodeI, bool, doctree.DimensionFlag) {
	lidx := dimension[0]
	// How accurate is the match.
	accuracy := doctree.DimensionLanguage
	switch v := n.(type) {
	case contentNodeIs:
		if len(v) == 0 {
			panic("empty contentNodeIs")
		}
		vv := v[lidx]
		if vv != nil {
			return vv, true, accuracy
		}
		return nil, false, 0
	case resourceSources:
		vv := v[lidx]
		if vv != nil {
			return vv, true, doctree.DimensionLanguage
		}
		if exact {
			return nil, false, 0
		}
		// For non content resources, pick the first match.
		for _, vv := range v {
			if vv != nil {
				if vv.isPage() {
					return nil, false, 0
				}
				return vv, true, 0
			}
		}
	case *resourceSource:
		if v.LangIndex() == lidx {
			return v, true, doctree.DimensionLanguage
		}
		if !v.isPage() && !exact {
			return v, true, 0
		}
	case *pageState:
		if v.s.languagei == lidx {
			return n, true, doctree.DimensionLanguage
		}
	default:
		panic(fmt.Sprintf("unknown type %T", n))
	}
	return nil, false, 0
}

func (s *ContentNodeShifter) ForEeachInDimension(n contentNodeI, d int, f func(contentNodeI) bool) {
	if d != doctree.DimensionLanguage.Index() {
		panic("only language dimension supported")
	}

	switch vv := n.(type) {
	case contentNodeIs:
		for _, v := range vv {
			if v != nil {
				if f(v) {
					return
				}
			}
		}
	default:
		f(vv)
	}
}

func (s *ContentNodeShifter) InsertInto(old, new contentNodeI, dimension doctree.Dimension) contentNodeI {
	langi := dimension[doctree.DimensionLanguage.Index()]
	switch vv := old.(type) {
	case *pageState:
		newp, ok := new.(*pageState)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		if vv.s.languagei == newp.s.languagei && newp.s.languagei == langi {
			return new
		}
		is := make(contentNodeIs, s.numLanguages)
		is[vv.s.languagei] = old
		is[langi] = new
		return is
	case contentNodeIs:
		vv[langi] = new
		return vv
	case resourceSources:
		vv[langi] = new.(*resourceSource)
		return vv
	case *resourceSource:
		newp, ok := new.(*resourceSource)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		if vv.LangIndex() == newp.LangIndex() && newp.LangIndex() == langi {
			return new
		}
		rs := make(resourceSources, s.numLanguages)
		rs[vv.LangIndex()] = vv
		rs[langi] = newp
		return rs

	default:
		panic(fmt.Sprintf("unknown type %T", old))
	}
}

func (s *ContentNodeShifter) Insert(old, new contentNodeI) contentNodeI {
	switch vv := old.(type) {
	case *pageState:
		newp, ok := new.(*pageState)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		if vv.s.languagei == newp.s.languagei {
			return new
		}
		is := make(contentNodeIs, s.numLanguages)
		is[newp.s.languagei] = new
		is[vv.s.languagei] = old
		return is
	case contentNodeIs:
		newp, ok := new.(*pageState)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		resource.MarkStale(vv[newp.s.languagei])
		vv[newp.s.languagei] = new
		return vv
	case *resourceSource:
		newp, ok := new.(*resourceSource)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		if vv.LangIndex() == newp.LangIndex() {
			return new
		}
		rs := make(resourceSources, s.numLanguages)
		rs[newp.LangIndex()] = newp
		rs[vv.LangIndex()] = vv
		return rs
	case resourceSources:
		newp, ok := new.(*resourceSource)
		if !ok {
			panic(fmt.Sprintf("unknown type %T", new))
		}
		resource.MarkStale(vv[newp.LangIndex()])
		vv[newp.LangIndex()] = newp
		return vv
	default:
		panic(fmt.Sprintf("unknown type %T", old))
	}
}
