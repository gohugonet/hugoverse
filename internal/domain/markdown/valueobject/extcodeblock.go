// Copyright 2024 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package valueobject

import (
	"bytes"
	"errors"
	"github.com/mdfriday/hugoverse/internal/domain/markdown"
	"strings"
	"sync"

	"github.com/mdfriday/hugoverse/pkg/herrors"
	htext "github.com/mdfriday/hugoverse/pkg/text"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type (
	codeBlocksExtension struct {
		markdown.Highlighter
	}
	htmlRenderer struct {
		markdown.Highlighter
	}
)

func NewCodeBlocksExt(highlighter markdown.Highlighter) goldmark.Extender {
	return &codeBlocksExtension{
		highlighter,
	}
}

func (e *codeBlocksExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&CodeBlockTransformer{}, 100),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newHTMLRenderer(e.Highlighter), 100),
	))
}

func newHTMLRenderer(highlighter markdown.Highlighter) renderer.NodeRenderer {
	r := &htmlRenderer{
		highlighter,
	}
	return r
}

func (r *htmlRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCodeBlock, r.renderCodeBlock)
}

func (r *htmlRenderer) renderCodeBlock(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	ctx := w.(*Context)

	if entering {
		return ast.WalkContinue, nil
	}

	n := node.(*codeBlock)
	lang := getLang(n.b, src)
	cbRenderer := ctx.RenderContext().GetRenderer(markdown.CodeBlockRendererType, lang)
	if cbRenderer == nil {
		cbRenderer = r.Highlighter // use default highlighter code block renderer
		//return ast.WalkStop, fmt.Errorf("no code renderer found for %q", lang)
	}

	ordinal := n.ordinal

	var buff bytes.Buffer

	l := n.b.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.b.Lines().At(i)
		buff.Write(line.Value(src))
	}

	s := htext.Chomp(buff.String())

	var info []byte
	if n.b.Info != nil {
		info = n.b.Info.Segment.Value(src)
	}

	attrtp := AttributesOwnerCodeBlockCustom
	if isd, ok := cbRenderer.(markdown.IsDefaultCodeBlockRendererProvider); (ok && isd.IsDefaultCodeBlockRenderer()) || GetChromaLexer(lang) != nil {
		// We say that this is a Chroma code block if it's the default code block renderer
		// or if the language is supported by Chroma.
		attrtp = AttributesOwnerCodeBlockChroma
	}

	// IsDefaultCodeBlockRendererProvider
	attrs, attrStr, err := getAttributes(n.b, info)
	if err != nil {
		return ast.WalkStop, &herrors.TextSegmentError{Err: err, Segment: attrStr}
	}
	cbctx := &codeBlockContext{
		page:             ctx.DocumentContext().Document,
		lang:             lang,
		code:             s,
		ordinal:          ordinal,
		AttributesHolder: NewAttr(attrs, attrtp),
	}

	cbctx.createPos = func() htext.Position {
		if resolver, ok := cbRenderer.(markdown.ElementPositionResolver); ok {
			return resolver.ResolvePosition(cbctx)
		}
		return htext.Position{
			Filename:     ctx.DocumentContext().Filename,
			LineNumber:   1,
			ColumnNumber: 1,
		}
	}

	cr := cbRenderer.(markdown.CodeBlockRenderer)

	err = cr.RenderCodeblock(
		ctx.RenderContext().Ctx,
		w,
		cbctx,
	)

	if err != nil {
		return ast.WalkContinue, herrors.NewFileErrorFromPos(err, cbctx.createPos())
	}

	return ast.WalkContinue, nil
}

type codeBlockContext struct {
	page    any
	lang    string
	code    string
	ordinal int

	// This is only used in error situations and is expensive to create,
	// to delay creation until needed.
	pos       htext.Position
	posInit   sync.Once
	createPos func() htext.Position

	*AttributesHolder
}

func (c *codeBlockContext) Page() any {
	return c.page
}

func (c *codeBlockContext) Type() string {
	return c.lang
}

func (c *codeBlockContext) Inner() string {
	return c.code
}

func (c *codeBlockContext) Ordinal() int {
	return c.ordinal
}

func (c *codeBlockContext) Position() htext.Position {
	c.posInit.Do(func() {
		c.pos = c.createPos()
	})
	return c.pos
}

func getLang(node *ast.FencedCodeBlock, src []byte) string {
	langWithAttributes := string(node.Language(src))
	lang, _, _ := strings.Cut(langWithAttributes, "{")
	return lang
}

func getAttributes(node *ast.FencedCodeBlock, infostr []byte) ([]ast.Attribute, string, error) {
	if node.Attributes() != nil {
		return node.Attributes(), "", nil
	}
	if infostr != nil {
		attrStartIdx := -1
		attrEndIdx := -1

		for idx, char := range infostr {
			if attrEndIdx == -1 && char == '{' {
				attrStartIdx = idx
			}
			if attrStartIdx != -1 && char == '}' {
				attrEndIdx = idx
				break
			}
		}

		if attrStartIdx != -1 && attrEndIdx != -1 {
			n := ast.NewTextBlock() // dummy node for storing attributes
			attrStr := infostr[attrStartIdx : attrEndIdx+1]
			if attrs, hasAttr := parser.ParseAttributes(text.NewReader(attrStr)); hasAttr {
				for _, attr := range attrs {
					n.SetAttribute(attr.Name, attr.Value)
				}
				return n.Attributes(), "", nil
			} else {
				return nil, string(attrStr), errors.New("failed to parse Markdown attributes; you may need to quote the values")
			}
		}
	}
	return nil, "", nil
}

// KindCodeBlock is the kind of an Hugo code block.
var KindCodeBlock = ast.NewNodeKind("HugoCodeBlock")

// Its raw contents are the plain text of the code block.
type codeBlock struct {
	ast.BaseBlock
	ordinal int
	b       *ast.FencedCodeBlock
}

func (*codeBlock) Kind() ast.NodeKind { return KindCodeBlock }

func (*codeBlock) IsRaw() bool { return true }

func (b *codeBlock) Dump(src []byte, level int) {
}

type CodeBlockTransformer struct{}

// Transform transforms the provided Markdown AST.
func (*CodeBlockTransformer) Transform(doc *ast.Document, reader text.Reader, pctx parser.Context) {
	var codeBlocks []*ast.FencedCodeBlock

	ast.Walk(doc, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			return ast.WalkContinue, nil
		}

		cb, ok := node.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		codeBlocks = append(codeBlocks, cb)

		return ast.WalkContinue, nil
	})

	for i, cb := range codeBlocks {
		b := &codeBlock{b: cb, ordinal: i}
		parent := cb.Parent()
		if parent != nil {
			parent.ReplaceChild(parent, cb, b)
		}
	}
}
