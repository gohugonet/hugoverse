package valueobject

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/pkg/types/hstring"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"strings"
)

type hookedRenderer struct {
	linkifyProtocol []byte
	html.Config
}

func (r *hookedRenderer) SetOption(name renderer.OptionName, value any) {
	r.Config.SetOption(name, value)
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs.
func (r *hookedRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *hookedRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Image)
	var lr markdown.LinkRenderer

	ctx, ok := w.(*Context)
	if ok {
		h := ctx.RenderContext().GetRenderer(markdown.ImageRendererType, nil)
		ok = h != nil
		if ok {
			lr = h.(markdown.LinkRenderer)
		}
	}

	if !ok {
		return r.renderImageDefault(w, source, node, entering)
	}

	if entering {
		// Store the current pos so we can capture the rendered text.
		ctx.PushPos(ctx.Buffer.Len())
		return ast.WalkContinue, nil
	}

	pos := ctx.PopPos()
	text := ctx.Buffer.Bytes()[pos:]
	ctx.Buffer.Truncate(pos)

	var (
		isBlock bool
		ordinal int
	)
	if b, ok := n.AttributeString(ImageAttrIsBlock); ok && b.(bool) {
		isBlock = true
	}
	if n, ok := n.AttributeString(ImageAttrOrdinal); ok {
		ordinal = n.(int)
	}

	// We use the attributes to signal from the parser whether the images is in
	// a block context or not.
	// We may find a better way to do that, but for now, we'll need to remove any
	// internal attributes before rendering.
	attrs := r.filterInternalAttributes(n.Attributes())

	err := lr.RenderLink(
		ctx.RenderContext().Ctx,
		w,
		imageLinkContext{
			linkContext: linkContext{
				page:             ctx.DocumentContext().Document,
				destination:      string(n.Destination),
				title:            string(n.Title),
				text:             hstring.RenderedString(text),
				plainText:        string(n.Text(source)),
				AttributesHolder: NewAttr(attrs, AttributesOwnerGeneral),
			},
			ordinal: ordinal,
			isBlock: isBlock,
		},
	)

	return ast.WalkContinue, err
}

const (
	internalAttrPrefix = "_h__"
)

func (r *hookedRenderer) filterInternalAttributes(attrs []ast.Attribute) []ast.Attribute {
	n := 0
	for _, x := range attrs {
		if !bytes.HasPrefix(x.Name, []byte(internalAttrPrefix)) {
			attrs[n] = x
			n++
		}
	}
	return attrs[:n]
}

// Fall back to the default Goldmark render funcs. Method below borrowed from:
// https://github.com/yuin/goldmark/blob/b611cd333a492416b56aa8d94b04a67bf0096ab2/renderer/html/html.go#L404
func (r *hookedRenderer) renderImageDefault(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")
	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n, source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		attrs := r.filterInternalAttributes(n.Attributes())
		RenderASTAttributes(w, attrs...)
	}
	if r.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

// Borrowed from Goldmark.
func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Text(source))
		} else if !c.HasChildren() {
			buf.Write(util.EscapeHTML(c.Text(source)))
			if t, ok := c.(*ast.Text); ok && t.SoftLineBreak() {
				buf.WriteByte('\n')
			}
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}

func (r *hookedRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	var lr markdown.LinkRenderer

	ctx, ok := w.(*Context)
	if ok {
		h := ctx.RenderContext().GetRenderer(markdown.LinkRendererType, nil)
		ok = h != nil
		if ok {
			lr = h.(markdown.LinkRenderer)
		}
	}

	if !ok {
		return r.renderLinkDefault(w, source, node, entering)
	}

	if entering {
		// Store the current pos so we can capture the rendered text.
		ctx.PushPos(ctx.Buffer.Len())
		return ast.WalkContinue, nil
	}

	pos := ctx.PopPos()
	text := ctx.Buffer.Bytes()[pos:]
	ctx.Buffer.Truncate(pos)

	err := lr.RenderLink(
		ctx.RenderContext().Ctx,
		w,
		linkContext{
			page:             ctx.DocumentContext().Document,
			destination:      string(n.Destination),
			title:            string(n.Title),
			text:             hstring.RenderedString(text),
			plainText:        string(n.Text(source)),
			AttributesHolder: EmptyAttr,
		},
	)

	return ast.WalkContinue, err
}

// Fall back to the default Goldmark render funcs. Method below borrowed from:
// https://github.com/yuin/goldmark/blob/b611cd333a492416b56aa8d94b04a67bf0096ab2/renderer/html/html.go#L404
func (r *hookedRenderer) renderLinkDefault(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		if r.Unsafe || !html.IsDangerousURL(n.Destination) {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

func (r *hookedRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.AutoLink)
	var lr markdown.LinkRenderer

	ctx, ok := w.(*Context)
	if ok {
		h := ctx.RenderContext().GetRenderer(markdown.LinkRendererType, nil)
		ok = h != nil
		if ok {
			lr = h.(markdown.LinkRenderer)
		}
	}

	if !ok {
		return r.renderAutoLinkDefault(w, source, node, entering)
	}

	url := string(r.autoLinkURL(n, source))
	label := string(n.Label(source))
	if n.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(url), "mailto:") {
		url = "mailto:" + url
	}

	err := lr.RenderLink(
		ctx.RenderContext().Ctx,
		w,
		linkContext{
			page:             ctx.DocumentContext().Document,
			destination:      url,
			text:             hstring.RenderedString(label),
			plainText:        label,
			AttributesHolder: EmptyAttr,
		},
	)

	return ast.WalkContinue, err
}

// Fall back to the default Goldmark render funcs. Method below borrowed from:
// https://github.com/yuin/goldmark/blob/5588d92a56fe1642791cf4aa8e9eae8227cfeecd/renderer/html/html.go#L439
func (r *hookedRenderer) renderAutoLinkDefault(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}

	_, _ = w.WriteString(`<a href="`)
	url := r.autoLinkURL(n, source)
	label := n.Label(source)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(url), []byte("mailto:")) {
		_, _ = w.WriteString("mailto:")
	}
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(url, false)))
	if n.Attributes() != nil {
		_ = w.WriteByte('"')
		html.RenderAttributes(w, n, html.LinkAttributeFilter)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString(`">`)
	}
	_, _ = w.Write(util.EscapeHTML(label))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

func (r *hookedRenderer) autoLinkURL(n *ast.AutoLink, source []byte) []byte {
	url := n.URL(source)
	if len(n.Protocol) > 0 && !bytes.Equal(n.Protocol, r.linkifyProtocol) {
		// The CommonMark spec says "http" is the correct protocol for links,
		// but this doesn't make much sense (the fact that they should care about the rendered output).
		// Note that n.Protocol is not set if protocol is provided by user.
		url = append(r.linkifyProtocol, url[len(n.Protocol):]...)
	}
	return url
}

func (r *hookedRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	var hr markdown.HeadingRenderer

	ctx, ok := w.(*Context)
	if ok {
		h := ctx.RenderContext().GetRenderer(markdown.HeadingRendererType, nil)
		ok = h != nil
		if ok {
			hr = h.(markdown.HeadingRenderer)
		}
	}

	if !ok {
		return r.renderHeadingDefault(w, source, node, entering)
	}

	if entering {
		// Store the current pos so we can capture the rendered text.
		ctx.PushPos(ctx.Buffer.Len())
		return ast.WalkContinue, nil
	}

	pos := ctx.PopPos()
	text := ctx.Buffer.Bytes()[pos:]
	ctx.Buffer.Truncate(pos)
	// All ast.Heading nodes are guaranteed to have an attribute called "id"
	// that is an array of bytes that encode a valid string.
	anchori, _ := n.AttributeString("id")
	anchor := anchori.([]byte)

	err := hr.RenderHeading(
		ctx.RenderContext().Ctx,
		w,
		headingContext{
			page:             ctx.DocumentContext().Document,
			level:            n.Level,
			anchor:           string(anchor),
			text:             hstring.RenderedString(text),
			plainText:        string(n.Text(source)),
			AttributesHolder: NewAttr(n.Attributes(), AttributesOwnerGeneral),
		},
	)

	return ast.WalkContinue, err
}

func (r *hookedRenderer) renderHeadingDefault(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			RenderASTAttributes(w, node.Attributes()...)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}
