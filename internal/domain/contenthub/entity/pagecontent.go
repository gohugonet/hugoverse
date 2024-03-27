package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

// The content related items on a Page.
type pageContent struct {
	truncated bool

	cmap *pageContentMap

	source rawPageContent
}

type pageContentMap struct {

	// If not, we can skip any pre-rendering of shortcodes.
	hasMarkdownShortcode bool

	// Indicates whether we must do placeholder replacements.
	hasNonMarkdownShortcode bool

	//  *shortcode, pageContentReplacement or pageparser.Item
	items []any
}

func (p *pageContentMap) AddBytes(item pageparser.Item) {
	p.items = append(p.items, item)
}

type rawPageContent struct {
	hasSummaryDivider bool

	// The AST of the parsed page. Contains information about:
	// shortcodes, front matter, summary indicators.
	parsed pageparser.Result

	// Returns the position in bytes after any front matter.
	posMainContent int

	// These are set if we're able to determine this from the source.
	posSummaryEnd int
	posBodyStart  int
}

type pageContentReplacement struct {
	val []byte

	source pageparser.Item
}

// returns the content to be processed by Goldmark or similar.
func (p pageContent) contentToRender(parsed pageparser.Result, pm *pageContentMap) []byte {
	source := parsed.Input()

	c := make([]byte, 0, len(source)+(len(source)/10))

	for _, it := range pm.items {
		switch v := it.(type) {
		case pageparser.Item:
			c = append(c, source[v.Pos():v.Pos()+len(v.Val(source))]...)
		case pageContentReplacement:
			c = append(c, v.val...)
		default:
			panic(fmt.Sprintf("unknown item type %T", it))
		}
	}

	return c
}
