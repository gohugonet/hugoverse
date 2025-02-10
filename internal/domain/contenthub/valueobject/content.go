package valueobject

import (
	"bytes"
	"github.com/mdfriday/hugoverse/internal/domain/markdown"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/media"
	"github.com/mdfriday/hugoverse/pkg/types"
	"html/template"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultTocConfig is the default ToC configuration.
var DefaultTocConfig = TocConfig{
	StartLevel: 2,
	EndLevel:   3,
	Ordered:    false,
}

type TocConfig struct {
	// Heading start level to include in the table of contents, starting
	// at h1 (inclusive).
	// <docsmeta>{ "identifiers": ["h1"] }</docsmeta>
	StartLevel int

	// Heading end level, inclusive, to include in the table of contents.
	// Default is 3, a value of -1 will include everything.
	EndLevel int

	// Whether to produce a ordered list or not.
	Ordered bool
}

type ContentSummary struct {
	Content             template.HTML
	Summary             template.HTML
	SummaryTruncated    bool
	TableOfContentsHTML template.HTML

	WordCount   int
	ReadingTime int

	MarkdownResult markdown.Result
}

func NewEmptyContentSummary() ContentSummary {
	return ContentSummary{
		Content:             "",
		Summary:             "",
		SummaryTruncated:    false,
		TableOfContentsHTML: "",

		MarkdownResult: nil,
	}
}

func (c *ContentSummary) IsSummaryEmpty() bool {
	return c.Summary == ""
}

func (c *ContentSummary) ExtractSummary(input []byte, mt media.Type) {
	in := string(input)

	res := c.extractSummaryFromHTML(mt, in, 70, containsCJK(in))
	sum := res.Summary()
	if sum != "" {
		c.Summary = helpers.BytesToHTML([]byte(sum))
		c.SummaryTruncated = res.Truncated()
		return
	}

	ts := c.trimShortHTML(input)
	c.Summary = helpers.BytesToHTML(input)
	c.SummaryTruncated = len(ts) < len(in)

	return
}

func (c *ContentSummary) extractSummaryFromHTML(mt media.Type, input string, numWords int, isCJK bool) (result *HtmlSummary) {
	result = &HtmlSummary{source: input}
	ptag := result.resolveParagraphTagAndSetWrapper(mt)

	if numWords <= 0 {
		return result
	}

	var count int

	countWord := func(word string) int {
		word = strings.TrimSpace(word)
		if len(word) == 0 {
			return 0
		}
		if isProbablyHTMLToken(word) {
			return 0
		}

		if isCJK {
			word = helpers.StripHTML(word)
			runeCount := utf8.RuneCountInString(word)
			if len(word) == runeCount {
				return 1
			} else {
				return runeCount
			}
		}

		return 1
	}

	high := len(input)
	if result.WrapperEnd.Low > 0 {
		high = result.WrapperEnd.Low
	}

	for j := result.WrapperStart.High; j < high; {
		s := input[j:]
		closingIndex := strings.Index(s, "</"+ptag.tagName+">")

		if closingIndex == -1 {
			break
		}

		s = s[:closingIndex]

		// Count the words in the current paragraph.
		var wi int

		for i, r := range s {
			if unicode.IsSpace(r) || (i+utf8.RuneLen(r) == len(s)) {
				word := s[wi:i]
				count += countWord(word)
				wi = i
				if count >= numWords {
					break
				}
			}
		}

		if count >= numWords {
			result.SummaryLowHigh = types.LowHigh[string]{
				Low:  result.WrapperStart.High,
				High: j + closingIndex + len(ptag.tagName) + 3,
			}
			return
		}

		j += closingIndex + len(ptag.tagName) + 2

	}

	result.SummaryLowHigh = types.LowHigh[string]{
		Low:  result.WrapperStart.High,
		High: high,
	}

	return
}

func containsCJK(s string) bool {
	re := regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]`)
	return re.MatchString(s)
}

// Avoid counting words that are most likely HTML tokens.
var (
	isProbablyHTMLTag      = regexp.MustCompile(`^<\/?[A-Za-z]+>?$`)
	isProablyHTMLAttribute = regexp.MustCompile(`^[A-Za-z]+=["']`)
)

func isProbablyHTMLToken(s string) bool {
	return s == ">" || isProbablyHTMLTag.MatchString(s) || isProablyHTMLAttribute.MatchString(s)
}

func (c *ContentSummary) trimShortHTML(input []byte) []byte {
	openingTag := []byte("<p>")
	closingTag := []byte("</p>")

	if bytes.Count(input, openingTag) == 1 {
		input = bytes.TrimSpace(input)
		if bytes.HasPrefix(input, openingTag) && bytes.HasSuffix(input, closingTag) {
			input = bytes.TrimPrefix(input, openingTag)
			input = bytes.TrimSuffix(input, closingTag)
			input = bytes.TrimSpace(input)
		}
	}
	return input
}
