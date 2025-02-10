package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/predicate"
	"strconv"
)

var pagePredicates = struct {
	KindPage         predicate.P[contenthub.Page]
	KindSection      predicate.P[contenthub.Page]
	KindHome         predicate.P[contenthub.Page]
	KindTerm         predicate.P[contenthub.Page]
	ShouldListLocal  predicate.P[contenthub.Page]
	ShouldListGlobal predicate.P[contenthub.Page]
	ShouldListAny    predicate.P[contenthub.Page]
	ShouldLink       predicate.P[contenthub.Page]
}{
	KindPage: func(p contenthub.Page) bool {
		return p.Kind() == valueobject.KindPage
	},
	KindSection: func(p contenthub.Page) bool {
		return p.Kind() == valueobject.KindSection
	},
	KindHome: func(p contenthub.Page) bool {
		return p.Kind() == valueobject.KindHome
	},
	KindTerm: func(p contenthub.Page) bool {
		return p.Kind() == valueobject.KindTerm
	},
	ShouldListLocal: func(p contenthub.Page) bool {
		return p.ShouldList(false)
	},
	ShouldListGlobal: func(p contenthub.Page) bool {
		return p.ShouldList(true)
	},
	ShouldListAny: func(p contenthub.Page) bool {
		return p.ShouldListAny()
	},
	ShouldLink: func(p contenthub.Page) bool {
		return !p.NoLink()
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
	Include predicate.P[contenthub.Page]
}

func (q pageMapQueryPagesBelowPath) Key() string {
	return q.Path + "/" + q.KeyPart
}
