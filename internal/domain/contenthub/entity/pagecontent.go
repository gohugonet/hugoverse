package entity

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

const (
	internalSummaryDividerBase = "HUGOMORE42"
)

var (
	internalSummaryDividerBaseBytes = []byte(internalSummaryDividerBase)
	InternalSummaryDividerPre       = []byte("\n\n" + internalSummaryDividerBase + "\n\n")
)

type Content struct {
	hasSummaryDivider bool
	summaryTruncated  bool

	rawSource []byte

	//  *shortcode, pageContentReplacement or pageparser.Item
	items []any
}

func NewContent(source []byte) *Content {
	return &Content{rawSource: source, items: make([]any, 0)}
}

func (c *Content) IsEmpty() bool {
	return c.rawSource == nil || len(c.rawSource) == 0
}

func (c *Content) SetSummaryDivider() {
	c.hasSummaryDivider = true
}

func (c *Content) Truncated() bool {
	return c.summaryTruncated
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
	return string(c.rawSource)
}

func (c *Content) PureContent() string {
	content := make([]byte, 0, len(c.rawSource)+(len(c.rawSource)/10))

	for _, it := range c.items {
		switch v := it.(type) {
		case pageparser.Item:
			content = append(content, c.rawSource[v.Pos():v.Pos()+len(v.Val(c.rawSource))]...)
		case valueobject.PageContentReplacement:
			content = append(content, v.Val...)
		case *valueobject.Shortcode:
			content = append(content, c.rawSource[v.Pos:v.Length]...)
		default:
			panic(fmt.Sprintf("unknown item type %T", v))
		}
	}

	return string(content)
}

func (c *Content) getShortCodes() []*valueobject.Shortcode {
	var res []*valueobject.Shortcode
	for _, item := range c.items {
		if s, ok := item.(*valueobject.Shortcode); ok {
			res = append(res, s)
		}
	}
	return res
}

// contentToRenderForItems returns the content to be processed by Goldmark or similar.
func (c *Content) contentToRender(ctx context.Context, source []byte,
	renderedShortcodes map[string]valueobject.ShortcodeRenderer) ([]byte, bool, error) {

	var hasVariants bool
	content := make([]byte, 0, len(source)+(len(source)/10))

	for _, it := range c.items {
		switch v := it.(type) {
		case pageparser.Item:
			content = append(content, source[v.Pos():v.Pos()+len(v.Val(source))]...)
		case valueobject.PageContentReplacement:
			content = append(content, v.Val...)
		case *valueobject.Shortcode:
			if !v.InsertPlaceholder() {
				// Insert the rendered shortcode.
				renderedShortcode, found := renderedShortcodes[v.Placeholder]
				if !found {
					// This should never happen.
					panic(fmt.Sprintf("rendered shortcode %q not found", v.Placeholder))
				}

				b, more, err := renderedShortcode.RenderShortcode(ctx)
				if err != nil {
					return nil, false, fmt.Errorf("failed to render shortcode: %w", err)
				}
				hasVariants = hasVariants || more
				content = append(content, []byte(b)...)

			} else {
				// Insert the placeholder so we can insert the content after
				// markdown processing.
				content = append(content, []byte(v.Placeholder)...)
			}
		default:
			panic(fmt.Sprintf("unknown item type %T", it))
		}
	}

	return content, hasVariants, nil
}

func splitUserDefinedSummaryAndContent(markup string, c []byte) (summary []byte, content []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("summary split failed: %s", r)
		}
	}()

	startDivider := bytes.Index(c, internalSummaryDividerBaseBytes)

	if startDivider == -1 {
		return
	}

	startTag := "p"
	switch markup {
	case "asciidocext":
		startTag = "div"
	}

	// Walk back and forward to the surrounding tags.
	start := bytes.LastIndex(c[:startDivider], []byte("<"+startTag))
	end := bytes.Index(c[startDivider:], []byte("</"+startTag))

	if start == -1 {
		start = startDivider
	} else {
		start = startDivider - (startDivider - start)
	}

	if end == -1 {
		end = startDivider + len(internalSummaryDividerBase)
	} else {
		end = startDivider + end + len(startTag) + 3
	}

	var addDiv bool

	switch markup {
	case "rst":
		addDiv = true
	}

	withoutDivider := append(c[:start], bytes.Trim(c[end:], "\n")...)

	if len(withoutDivider) > 0 {
		summary = bytes.TrimSpace(withoutDivider[:start])
	}

	if addDiv {
		// For the rst
		summary = append(append([]byte(nil), summary...), []byte("</div>")...)
	}

	if err != nil {
		return
	}

	content = bytes.TrimSpace(withoutDivider)

	return
}
