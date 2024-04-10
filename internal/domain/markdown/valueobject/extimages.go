package valueobject

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type (
	imagesExtension struct {
		wrapStandAloneImageWithinParagraph bool
	}
)

const (
	// Used to signal to the rendering step that an image is used in a block context.
	// Dont's change this; the prefix must match the internalAttrPrefix in the root goldmark package.

	ImageAttrIsBlock = "_h__isBlock"
	ImageAttrOrdinal = "_h__ordinal"
)

func NewImagesExt(wrapStandAloneImageWithinParagraph bool) goldmark.Extender {
	return &imagesExtension{wrapStandAloneImageWithinParagraph: wrapStandAloneImageWithinParagraph}
}

func (e *imagesExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&ImagesTransformer{wrapStandAloneImageWithinParagraph: e.wrapStandAloneImageWithinParagraph}, 300),
		),
	)
}

type ImagesTransformer struct {
	wrapStandAloneImageWithinParagraph bool
}

// Transform transforms the provided Markdown AST.
func (t *ImagesTransformer) Transform(doc *ast.Document, reader text.Reader, pctx parser.Context) {
	var ordinal int
	ast.Walk(doc, func(node ast.Node, enter bool) (ast.WalkStatus, error) {
		if !enter {
			return ast.WalkContinue, nil
		}

		if n, ok := node.(*ast.Image); ok {
			parent := n.Parent()
			n.SetAttributeString(ImageAttrOrdinal, ordinal)
			ordinal++

			if !t.wrapStandAloneImageWithinParagraph {
				isBlock := parent.ChildCount() == 1
				if isBlock {
					n.SetAttributeString(ImageAttrIsBlock, true)
				}

				if isBlock && parent.Kind() == ast.KindParagraph {
					for _, attr := range parent.Attributes() {
						// Transfer any attribute set down to the image.
						// Image elements does not support attributes on its own,
						// so it's safe to just set without checking first.
						n.SetAttribute(attr.Name, attr.Value)
					}
					grandParent := parent.Parent()
					grandParent.ReplaceChild(grandParent, parent, n)
				}
			}

		}

		return ast.WalkContinue, nil
	})
}
