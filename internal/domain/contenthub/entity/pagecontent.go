package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
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

	rawSource []byte

	//  *shortcode, pageContentReplacement or pageparser.Item
	items []any

	// Temporary storage of placeholders mapped to their content.
	// These are shortcodes etc. Some of these will need to be replaced
	// after any markup is rendered, so they share a common prefix.
	contentPlaceholders map[string]shortcodeRenderer
}

func NewContent(source []byte) *Content {
	return &Content{rawSource: source}
}

func (c *Content) SetSummaryDivider() {
	c.hasSummaryDivider = true
}

func (c *Content) SetSummaryTruncated() {
	c.summaryTruncated = true
}

func (c *Content) AddReplacement(val []byte, source pageparser.Item) {
	c.items = append(c.items, valueobject.PageContentReplacement{Val: val, Source: source})
}

func (c *Content) AddShortcode(s *valueobject.Shortcode) {
	c.items = append(c.items, s)
}

func (c *Content) AddItems(item pageparser.Item) {
	c.items = append(c.items, item)
}

func (c *Content) RawContent() string {
	// TODO, put empty here for new page builder

	return ""
}
