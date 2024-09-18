package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/predicate"
	"strconv"
)

var pagePredicates = struct {
	KindPage         predicate.P[*Page]
	KindSection      predicate.P[*Page]
	KindHome         predicate.P[*Page]
	KindTerm         predicate.P[*Page]
	ShouldListLocal  predicate.P[*Page]
	ShouldListGlobal predicate.P[*Page]
	ShouldListAny    predicate.P[*Page]
	ShouldLink       predicate.P[contenthub.Page]
}{
	KindPage: func(p *Page) bool {
		return p.Kind() == valueobject.KindPage
	},
	KindSection: func(p *Page) bool {
		return p.Kind() == valueobject.KindSection
	},
	KindHome: func(p *Page) bool {
		return p.Kind() == valueobject.KindHome
	},
	KindTerm: func(p *Page) bool {
		return p.Kind() == valueobject.KindTerm
	},
	ShouldListLocal: func(p *Page) bool {
		return p.Meta.shouldList(false)
	},
	ShouldListGlobal: func(p *Page) bool {
		return p.Meta.shouldList(true)
	},
	ShouldListAny: func(p *Page) bool {
		return p.Meta.shouldListAny()
	},
	ShouldLink: func(p contenthub.Page) bool {
		return !p.(*Page).Meta.noLink()
	},
}

type pageMapQueryPagesInSection struct {
	pageMapQueryPagesBelowPath

	Recursive   bool
	IncludeSelf bool
	Index       int
}

func (q pageMapQueryPagesInSection) Key() string {
	return "gagesInSection" + "/" + q.pageMapQueryPagesBelowPath.Key() + "/" + strconv.FormatBool(q.Recursive) +
		"/" + strconv.Itoa(q.Index) + "/" + strconv.FormatBool(q.IncludeSelf)
}

type pageMapQueryPagesBelowPath struct {
	Path string

	// Additional identifier for this query.
	// Used as part of the cache key.
	KeyPart string

	// Page inclusion filter.
	// May be nil.
	Include predicate.P[*Page]
}

func (q pageMapQueryPagesBelowPath) Key() string {
	return q.Path + "/" + q.KeyPart
}
