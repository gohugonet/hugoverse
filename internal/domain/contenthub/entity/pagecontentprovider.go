package entity

import "C"
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/gohugonet/hugoverse/pkg/text"
	goTemplate "html/template"
	"reflect"
	"regexp"
)

const (
	innerNewlineRegexp = "\n"
	innerCleanupRegexp = `\A<p>(.*)</p>\n\z`
	innerCleanupExpand = "$1"
)

type ContentProvider struct {
	source  *Source
	content *Content

	page *Page

	cache *Cache
	f     output.Format

	// Temporary storage of placeholders mapped to their content.
	// These are shortcodes etc. Some of these will need to be replaced
	// after any markup is rendered, so they share a common prefix.
	contentPlaceholders map[string]valueobject.ShortcodeRenderer

	converter   contenthub.Converter
	templateSvc contenthub.Template

	log loggers.Logger
}

func (c *ContentProvider) cacheKey() string {
	return c.source.sourceKey() + "/" + c.f.Name
}

func (c *ContentProvider) Summary() goTemplate.HTML {
	cs, err := c.ContentSummary()
	if err != nil {
		c.log.Errorln(err)
		return goTemplate.HTML("")
	}

	return cs.Summary
}

func (c *ContentProvider) TableOfContents() goTemplate.HTML {
	cs, err := c.ContentSummary()
	if err != nil {
		c.log.Errorln(err)
		return goTemplate.HTML("")
	}

	return cs.TableOfContentsHTML
}

func (c *ContentProvider) Content() (any, error) {
	cs, err := c.ContentSummary()
	return cs.Content, err
}

func (c *ContentProvider) ContentSummary() (valueobject.ContentSummary, error) {
	v, err := c.cache.CacheContentRendered.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[valueobject.ContentSummary], error) {
		if c.content == nil {
			return &stale.Value[valueobject.ContentSummary]{
				Value: valueobject.NewEmptyContentSummary(),
				IsStaleFunc: func() bool {
					return c.source.IsStale()
				},
			}, nil
		}

		renderers, err := c.shortcodes()
		if err != nil {
			return nil, err
		}
		c.contentPlaceholders = renderers

		contentToRender, err := c.contentToRender()
		if err != nil {
			return nil, err
		}

		res, err := c.converter.Convert(markdown.RenderContext{
			Ctx:       context.Background(),
			Src:       contentToRender,
			RenderTOC: true,
			GetRenderer: func(t markdown.RendererType, id any) any {
				return nil
			},
		})
		if err != nil {
			return nil, err
		}

		b := res.Bytes()
		v := valueobject.NewEmptyContentSummary()
		v.TableOfContentsHTML = res.TableOfContents().ToHTML(
			valueobject.DefaultTocConfig.StartLevel,
			valueobject.DefaultTocConfig.EndLevel,
			valueobject.DefaultTocConfig.Ordered)

		// There are one or more replacement tokens to be replaced.
		var hasShortcodeVariants bool
		tokenHandler := func(ctx context.Context, token string) ([]byte, error) {
			if token == valueobject.TocShortcodePlaceholder {
				return []byte(v.TableOfContentsHTML), nil
			}
			renderer, found := c.contentPlaceholders[token]
			if found {
				repl, more, err := renderer.RenderShortcode(ctx)
				if err != nil {
					return nil, err
				}
				hasShortcodeVariants = hasShortcodeVariants || more
				return repl, nil
			}
			// This should never happen.
			panic(fmt.Errorf("unknown shortcode token %q (number of tokens: %d)", token, len(c.contentPlaceholders)))
		}

		b, err = expandShortcodeTokens(context.Background(), b, tokenHandler)
		if err != nil {
			return nil, err
		}

		if c.content.hasSummaryDivider {
			summary, content, err := splitUserDefinedSummaryAndContent("markdown", b)
			if err != nil {
				c.log.Errorf("Failed to set user defined summary for page %q: %s", c.source.File.FileName(), err)
			} else {
				b = content
				v.Summary = helpers.BytesToHTML(summary)
			}
		}

		v.SummaryTruncated = c.content.summaryTruncated
		v.Content = helpers.BytesToHTML(b)

		return &stale.Value[valueobject.ContentSummary]{
			Value: v,
			IsStaleFunc: func() bool {
				return c.source.IsStale()
			},
		}, nil
	})
	if err != nil {
		return valueobject.NewEmptyContentSummary(), err
	}

	return v.Value, nil
}

func (c *ContentProvider) contentToRender() ([]byte, error) {
	v, err := c.cache.CacheContentToRender.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[[]byte], error) {
		source, err := c.source.contentSource()
		if err != nil {
			return nil, err
		}
		contentToRender, _, err := c.content.contentToRender(context.Background(), source, c.contentPlaceholders)
		if err != nil {
			return nil, err
		}

		return &stale.Value[[]byte]{
			Value: contentToRender,
			IsStaleFunc: func() bool {
				return c.source.IsStale()
			},
		}, nil
	})

	if err != nil {
		return nil, err
	}

	return v.Value, nil
}

func (c *ContentProvider) shortcodes() (map[string]valueobject.ShortcodeRenderer, error) {
	v, err := c.cache.CacheContentShortcodes.GetOrCreate(c.cacheKey(),
		func(string) (*stale.Value[map[string]valueobject.ShortcodeRenderer], error) {
			renderers, err := c.prepareShortcodesForPage()
			if err != nil {
				return nil, err
			}
			return &stale.Value[map[string]valueobject.ShortcodeRenderer]{
				Value: renderers,
				IsStaleFunc: func() bool {
					return c.source.IsStale()
				},
			}, nil
		})
	if err != nil {
		return make(map[string]valueobject.ShortcodeRenderer), err
	}

	return v.Value, nil
}

func (c *ContentProvider) prepareShortcodesForPage() (map[string]valueobject.ShortcodeRenderer, error) {
	rendered := make(map[string]valueobject.ShortcodeRenderer)

	for _, v := range c.content.getShortCodes() {
		s, err := c.prepareShortcode(v, nil, 0)
		if err != nil {
			return nil, err
		}
		rendered[v.Placeholder] = s

	}

	return rendered, nil
}

func (c *ContentProvider) prepareShortcode(sc *valueobject.Shortcode, parent *ShortcodeWithPage, level int) (valueobject.ShortcodeRenderer, error) {
	toParseErr := func(err error) error {
		source := c.content.rawSource
		return c.parseError(fmt.Errorf("failed to render shortcode %q: %w", sc.Name, err), source, sc.Pos)
	}

	// Allow the caller to delay the rendering of the shortcode if needed.
	var fn valueobject.ShortcodeRenderFunc = func(ctx context.Context) ([]byte, bool, error) {
		r, err := c.doRenderShortcode(sc, parent, level)
		if err != nil {
			return nil, false, toParseErr(err)
		}
		b, hasVariants, err := r.RenderShortcode(ctx)
		if err != nil {
			return nil, false, toParseErr(err)
		}
		return b, hasVariants, nil
	}

	return fn, nil
}

func (c *ContentProvider) parseError(err error, input []byte, offset int) error {
	pos := valueobject.PosFromInput("", input, offset)
	return herrors.NewFileErrorFromName(err, c.source.File.Filename()).UpdatePosition(pos)
}

func (c *ContentProvider) doRenderShortcode(sc *valueobject.Shortcode, parent *ShortcodeWithPage, level int) (valueobject.ShortcodeRenderer, error) {
	var tmpl template.Preparer

	// Tracks whether this shortcode or any of its children has template variations
	// in other languages or output formats. We are currently only interested in
	// the output formats, so we may get some false positives -- we
	// should improve on that.
	var hasVariants bool

	tplVariants := template.Variants{
		Language:     c.source.PageLanguage(),
		OutputFormat: c.f,
	}

	if sc.IsInline {
		c.log.Warnf("Rendering inline shortcode not supported yet: %q", sc.Name)
		return valueobject.ZeroShortcode, nil
	} else {
		var found, more bool
		tmpl, found, more = c.templateSvc.LookupVariant(sc.Name, tplVariants)
		if !found {
			c.log.Errorf("Unable to locate template for shortcode %q in page %q", sc.Name, c.source.File.Path().Path())
			return valueobject.ZeroShortcode, nil
		}
		hasVariants = hasVariants || more
	}

	data := &ShortcodeWithPage{Ordinal: sc.Ordinal, posOffset: sc.Pos, indentation: sc.Indentation,
		Params: sc.Params, Page: c.page, Parent: parent, Name: sc.Name}

	if sc.Params != nil {
		data.IsNamedParams = reflect.TypeOf(sc.Params).Kind() == reflect.Map
	}

	if len(sc.Inner) > 0 {
		var inner string
		for _, innerData := range sc.Inner {
			switch innerData := innerData.(type) {
			case string:
				inner += innerData
			case *valueobject.Shortcode:
				s, err := c.prepareShortcode(innerData, data, level+1)
				if err != nil {
					return valueobject.ZeroShortcode, err
				}
				ss, more, err := s.RenderShortcodeString(context.Background())
				hasVariants = hasVariants || more
				if err != nil {
					return valueobject.ZeroShortcode, err
				}
				inner += ss
			default:
				c.log.Errorf("Illegal state on shortcode rendering of %q in page %q. Illegal type in inner data: %s ",
					sc.Name, c.source.File.Path().Path(), reflect.TypeOf(innerData))
				return valueobject.ZeroShortcode, nil
			}
		}

		// Pre Hugo 0.55 this was the behavior even for the outer-most
		// shortcode.
		if sc.DoMarkup && level > 0 {
			var err error
			b, err := c.converter.Convert(
				markdown.RenderContext{
					Ctx:       context.Background(),
					Src:       []byte(inner),
					RenderTOC: false,
					GetRenderer: func(t markdown.RendererType, id any) any {
						return nil
					},
				})
			if err != nil {
				return valueobject.ZeroShortcode, err
			}

			newInner := b.Bytes()

			// If the type is “” (unknown) or “markdown”, we assume the markdown
			// generation has been performed. Given the input: `a line`, markdown
			// specifies the HTML `<p>a line</p>\n`. When dealing with documents as a
			// whole, this is OK. When dealing with an `{{ .Inner }}` block in Hugo,
			// this is not so good. This code does two things:
			//
			// 1.  Check to see if inner has a newline in it. If so, the Inner data is
			//     unchanged.
			// 2   If inner does not have a newline, strip the wrapping <p> block and
			//     the newline.
			if match, _ := regexp.MatchString(innerNewlineRegexp, inner); !match {
				cleaner, err := regexp.Compile(innerCleanupRegexp)

				if err == nil {
					newInner = cleaner.ReplaceAll(newInner, []byte(innerCleanupExpand))
				}
			}

			// TODO(bep) we may have plain text inner templates.
			data.Inner = goTemplate.HTML(newInner)
		} else {
			data.Inner = goTemplate.HTML(inner)
		}

	}

	result, err := c.renderShortcodeWithPage(tmpl, data)

	if err != nil && sc.IsInline {
		fe := herrors.NewFileErrorFromName(err, c.source.File.Filename())
		pos := fe.Position()
		pos.LineNumber += c.page.source.posOffset(sc.Pos).LineNumber
		fe = fe.UpdatePosition(pos)
		return valueobject.ZeroShortcode, fe
	}

	if len(sc.Inner) == 0 && len(sc.Indentation) > 0 {
		b := bp.GetBuffer()
		i := 0
		text.VisitLinesAfter(result, func(line string) {
			// The first line is correctly indented.
			if i > 0 {
				b.WriteString(sc.Indentation)
			}
			i++
			b.WriteString(line)
		})

		result = b.String()
		bp.PutBuffer(b)
	}

	return valueobject.NewPrerenderedShortcode(result, hasVariants), nil
}

func (c *ContentProvider) renderShortcodeWithPage(tmpl template.Preparer, data *ShortcodeWithPage) (string, error) {
	buffer := bp.GetBuffer()
	defer bp.PutBuffer(buffer)

	err := c.templateSvc.ExecuteWithContext(context.Background(), tmpl, buffer, data)
	if err != nil {
		return "", fmt.Errorf("failed to process shortcode: %w", err)
	}
	return buffer.String(), nil
}

// Replace prefixed shortcode tokens with the real content.
// Note: This function will rewrite the input slice.
func expandShortcodeTokens(
	ctx context.Context,
	source []byte,
	tokenHandler func(ctx context.Context, token string) ([]byte, error),
) ([]byte, error) {
	start := 0

	pre := []byte(valueobject.ShortcodePlaceholderPrefix)
	post := []byte("HBHB")
	pStart := []byte("<p>")
	pEnd := []byte("</p>")

	k := bytes.Index(source[start:], pre)

	for k != -1 {
		j := start + k
		postIdx := bytes.Index(source[j:], post)
		if postIdx < 0 {
			// this should never happen, but let the caller decide to panic or not
			return nil, errors.New("illegal state in content; shortcode token missing end delim")
		}

		end := j + postIdx + 4
		key := string(source[j:end])
		newVal, err := tokenHandler(ctx, key)
		if err != nil {
			return nil, err
		}

		// Issue #1148: Check for wrapping p-tags <p>
		if j >= 3 && bytes.Equal(source[j-3:j], pStart) {
			if (k+4) < len(source) && bytes.Equal(source[end:end+4], pEnd) {
				j -= 3
				end += 4
			}
		}

		// This and other cool slice tricks: https://github.com/golang/go/wiki/SliceTricks
		source = append(source[:j], append(newVal, source[end:]...)...)
		start = j
		k = bytes.Index(source[start:], pre)

	}

	return source, nil
}
