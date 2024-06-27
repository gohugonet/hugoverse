package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

const (
	internalSummaryDividerBase = "HUGOMORE42"
)

var (
	internalSummaryDividerBaseBytes = []byte(internalSummaryDividerBase)
	internalSummaryDividerPre       = []byte("\n\n" + internalSummaryDividerBase + "\n\n")
)

type Content struct {
	hasSummaryDivider bool
	summaryTruncated  bool

	source []byte

	//  *shortcode, pageContentReplacement or pageparser.Item
	items []any
}

func (c *Content) AddReplacement(val []byte, source pageparser.Item) {
	c.items = append(c.items, valueobject.PageContentReplacement{Val: val, Source: source})
}

func (c *Content) AddShortcode(s *valueobject.Shortcode) {
	c.items = append(c.items, s)
}

func (c *Content) bytesHandler(item pageparser.Item) error {
	c.items = append(c.items, item)
	return nil
}

func (c *Content) summaryHandler(it pageparser.Item, iter *pageparser.Iterator) error {
	posBody := -1
	f := func(item pageparser.Item) bool {
		if posBody == -1 && !item.IsDone() {
			posBody = item.Pos()
		}

		if item.IsNonWhitespace(c.source) {
			c.summaryTruncated = true

			// Done
			return false
		}
		return true
	}
	iter.PeekWalk(f)

	c.hasSummaryDivider = true

	// The content may be rendered by Goldmark or similar,
	// and we need to track the summary.
	c.AddReplacement(internalSummaryDividerPre, it)

	return nil
}
