package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

const (
	internalSummaryDividerBase = "HUGOMORE42"
)

var (
	InternalSummaryDividerBaseBytes = []byte(internalSummaryDividerBase)
	InternalSummaryDividerPre       = []byte("\n\n" + internalSummaryDividerBase + "\n\n")
)

type Content struct {
	hasSummaryDivider bool
	summaryTruncated  bool

	//  *shortcode, pageContentReplacement or pageparser.Item
	items []any
}

func NewContent() *Content {
	return &Content{}
}

func (c *Content) SetSummaryDivider() {
	c.hasSummaryDivider = true
}

func (c *Content) SetSummaryTruncated() {
	c.summaryTruncated = true
}

func (c *Content) AddReplacement(val []byte, source pageparser.Item) {
	c.items = append(c.items, PageContentReplacement{Val: val, Source: source})
}

func (c *Content) AddShortcode(s *Shortcode) {
	c.items = append(c.items, s)
}

func (c *Content) AddItems(item pageparser.Item) {
	c.items = append(c.items, item)
}

func (c *Content) RawContent() string {
	// TODO, put empty here for new page builder

	return ""
}
