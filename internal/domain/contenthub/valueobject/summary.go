package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/types"
	"regexp"
	"strings"
)

type HtmlSummary struct {
	source         string
	SummaryLowHigh types.LowHigh[string]
	SummaryEndTag  types.LowHigh[string]
	WrapperStart   types.LowHigh[string]
	WrapperEnd     types.LowHigh[string]
	Divider        types.LowHigh[string]
}

func (s *HtmlSummary) wrap(ss string) string {
	if s.WrapperStart.IsZero() {
		return ss
	}
	return s.source[s.WrapperStart.Low:s.WrapperStart.High] + ss + s.source[s.WrapperEnd.Low:s.WrapperEnd.High]
}

func (s *HtmlSummary) wrapLeft(ss string) string {
	if s.WrapperStart.IsZero() {
		return ss
	}

	return s.source[s.WrapperStart.Low:s.WrapperStart.High] + ss
}

func (s *HtmlSummary) Value(l types.LowHigh[string]) string {
	return s.source[l.Low:l.High]
}

func (s *HtmlSummary) trimSpace(ss string) string {
	return strings.TrimSpace(ss)
}

func (s *HtmlSummary) Content() string {
	if s.Divider.IsZero() {
		return s.source
	}
	ss := s.source[:s.Divider.Low]
	ss += s.source[s.Divider.High:]
	return s.trimSpace(ss)
}

func (s *HtmlSummary) Summary() string {
	if s.Divider.IsZero() {
		return s.trimSpace(s.wrap(s.Value(s.SummaryLowHigh)))
	}
	ss := s.source[s.SummaryLowHigh.Low:s.Divider.Low]
	if s.SummaryLowHigh.High > s.Divider.High {
		ss += s.source[s.Divider.High:s.SummaryLowHigh.High]
	}
	if !s.SummaryEndTag.IsZero() {
		ss += s.Value(s.SummaryEndTag)
	}
	return s.trimSpace(s.wrap(ss))
}

func (s *HtmlSummary) ContentWithoutSummary() string {
	if s.Divider.IsZero() {
		if s.SummaryLowHigh.Low == s.WrapperStart.High && s.SummaryLowHigh.High == s.WrapperEnd.Low {
			return ""
		}
		return s.trimSpace(s.wrapLeft(s.source[s.SummaryLowHigh.High:]))
	}
	if s.SummaryEndTag.IsZero() {
		return s.trimSpace(s.wrapLeft(s.source[s.Divider.High:]))
	}
	return s.trimSpace(s.wrapLeft(s.source[s.SummaryEndTag.High:]))
}

func (s *HtmlSummary) Truncated() bool {
	return s.SummaryLowHigh.High < len(s.source)
}

func (s *HtmlSummary) resolveParagraphTagAndSetWrapper(mt media.Type) tagReStartEnd {
	ptag := startEndP

	switch mt.SubType {
	case DefaultContentTypes.AsciiDoc.SubType:
		ptag = startEndDiv
	case DefaultContentTypes.ReStructuredText.SubType:
		const markerStart = "<div class=\"document\">"
		const markerEnd = "</div>"
		i1 := strings.Index(s.source, markerStart)
		i2 := strings.LastIndex(s.source, markerEnd)
		if i1 > -1 && i2 > -1 {
			s.WrapperStart = types.LowHigh[string]{Low: 0, High: i1 + len(markerStart)}
			s.WrapperEnd = types.LowHigh[string]{Low: i2, High: len(s.source)}
		}
	}
	return ptag
}

var (
	pOrDiv = regexp.MustCompile(`<p[^>]?>|<div[^>]?>$`)

	startEndDiv = tagReStartEnd{
		startEndOfString: regexp.MustCompile(`<div[^>]*?>$`),
		endEndOfString:   regexp.MustCompile(`</div>$`),
		tagName:          "div",
	}

	startEndP = tagReStartEnd{
		startEndOfString: regexp.MustCompile(`<p[^>]*?>$`),
		endEndOfString:   regexp.MustCompile(`</p>$`),
		tagName:          "p",
	}
)

type tagReStartEnd struct {
	startEndOfString *regexp.Regexp
	endEndOfString   *regexp.Regexp
	tagName          string
}
