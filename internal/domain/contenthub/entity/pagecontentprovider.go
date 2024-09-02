package entity

import "C"
import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/gohugonet/hugoverse/pkg/text"
	"reflect"
	"regexp"
)

type ContentProvider struct {
	source  *Source
	content *Content

	cache *Cache
	f     output.Format

	contentToRender []byte
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

func (c *ContentProvider) Content() (any, error) {
	v, err := c.cache.CacheContentRendered.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[valueobject.ContentSummary], error) {

		return nil, nil
	})
	if err != nil {
		return valueobject.ContentSummary{}, err
	}

	return v.Value, nil
}

func (c *ContentProvider) toc() (valueobject.ContentToC, error) {
	v, err := c.cache.CacheContentToCs.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[valueobject.ContentToC], error) {
		return nil, nil
	})
	if err != nil {
		return valueobject.ContentToC{}, err
	}

	return v.Value, nil
}

func (c *ContentProvider) shortcodes() (map[string]valueobject.ShortcodeRenderer, error) {
	v, err := c.cache.CacheContentShortcodes.GetOrCreate(c.cacheKey(),
		func(string) (*stale.Value[map[string]valueobject.ShortcodeRenderer], error) {

			return nil, nil
		})
	if err != nil {
		return make(map[string]valueobject.ShortcodeRenderer), err
	}

	return v.Value, nil
}

func (c *ContentProvider) prepareShortcodesForPage() error {
	rendered := make(map[string]valueobject.ShortcodeRenderer)

	for _, v := range c.content.getShortCodes() {
		s, err := c.prepareShortcode(v)
		if err != nil {
			return err
		}
		rendered[v.Placeholder] = s

	}

	c.contentPlaceholders = rendered

	return nil
}

func (c *ContentProvider) prepareShortcode(sc *valueobject.Shortcode) (valueobject.ShortcodeRenderer, error) {
	toParseErr := func(err error) error {
		source := c.content.rawSource
		return c.parseError(fmt.Errorf("failed to render shortcode %q: %w", sc.Name, err), source, sc.Pos)
	}

	// Allow the caller to delay the rendering of the shortcode if needed.
	var fn valueobject.ShortcodeRenderFunc = func(ctx context.Context) ([]byte, bool, error) {
		r, err := c.doRenderShortcode(sc)
		if err != nil {
			return nil, false, toParseErr(err)
		}
		b, hasVariants, err := r.renderShortcode(ctx)
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

func (c *ContentProvider) doRenderShortcode(sc *valueobject.Shortcode) (valueobject.ShortcodeRenderer, error) {
	var tmpl template.Preparer

	// Tracks whether this shortcode or any of its children has template variations
	// in other languages or output formats. We are currently only interested in
	// the output formats, so we may get some false positives -- we
	// should improve on that.
	var hasVariants bool

	tplVariants := template.Variants{
		Language:     c.source.Language(),
		OutputFormat: c.f,
	}

	if sc.IsInline {
		c.log.Warnf("Rendering inline shortcode not supported yet: %q", sc.Name)
		return valueobject.ZeroShortcode, nil
	} else {
		var found, more bool
		tmpl, found, more = c.templateSvc.LookupVariant(sc.Name, tplVariants)
		if !found {
			c.log.Errorf("Unable to locate template for shortcode %q in page %q", sc.Name, c.source.File.Path())
			return valueobject.ZeroShortcode, nil
		}
		hasVariants = hasVariants || more
	}

	data := &ShortcodeWithPage{Ordinal: sc.ordinal, posOffset: sc.pos, indentation: sc.indentation, Params: sc.params, Page: newPageForShortcode(p), Parent: parent, Name: sc.name}
	if sc.params != nil {
		data.IsNamedParams = reflect.TypeOf(sc.params).Kind() == reflect.Map
	}

	if len(sc.Inner) > 0 {
		var inner string
		for _, innerData := range sc.inner {
			switch innerData := innerData.(type) {
			case string:
				inner += innerData
			case *shortcode:
				s, err := prepareShortcode(ctx, level+1, s, tplVariants, innerData, data, p, isRenderString)
				if err != nil {
					return zeroShortcode, err
				}
				ss, more, err := s.renderShortcodeString(ctx)
				hasVariants = hasVariants || more
				if err != nil {
					return zeroShortcode, err
				}
				inner += ss
			default:
				s.Log.Errorf("Illegal state on shortcode rendering of %q in page %q. Illegal type in inner data: %s ",
					sc.name, p.File().Path(), reflect.TypeOf(innerData))
				return zeroShortcode, nil
			}
		}

		// Pre Hugo 0.55 this was the behavior even for the outer-most
		// shortcode.
		if sc.doMarkup && (level > 0 || sc.configVersion() == 1) {
			var err error
			b, err := p.pageOutput.contentRenderer.ParseAndRenderContent(ctx, []byte(inner), false)
			if err != nil {
				return zeroShortcode, err
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
			switch p.m.pageConfig.Markup {
			case "", "markdown":
				if match, _ := regexp.MatchString(innerNewlineRegexp, inner); !match {
					cleaner, err := regexp.Compile(innerCleanupRegexp)

					if err == nil {
						newInner = cleaner.ReplaceAll(newInner, []byte(innerCleanupExpand))
					}
				}
			}

			// TODO(bep) we may have plain text inner templates.
			data.Inner = template.HTML(newInner)
		} else {
			data.Inner = template.HTML(inner)
		}

	}

	result, err := renderShortcodeWithPage(ctx, s.Tmpl(), tmpl, data)

	if err != nil && sc.isInline {
		fe := herrors.NewFileErrorFromName(err, p.File().Filename())
		pos := fe.Position()
		pos.LineNumber += p.posOffset(sc.pos).LineNumber
		fe = fe.UpdatePosition(pos)
		return zeroShortcode, fe
	}

	if len(sc.inner) == 0 && len(sc.indentation) > 0 {
		b := bp.GetBuffer()
		i := 0
		text.VisitLinesAfter(result, func(line string) {
			// The first line is correctly indented.
			if i > 0 {
				b.WriteString(sc.indentation)
			}
			i++
			b.WriteString(line)
		})

		result = b.String()
		bp.PutBuffer(b)
	}

	return prerenderedShortcode{s: result, hasVariants: hasVariants}, err
}
