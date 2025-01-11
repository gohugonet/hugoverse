package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/yuin/goldmark/ast"
	"strings"
)

type ParserResult struct {
	doc ast.Node
	toc markdown.TocFragments
	src []byte
}

func NewParserResult(doc ast.Node, toc markdown.TocFragments, src []byte) *ParserResult {
	return &ParserResult{
		doc: doc,
		toc: toc,
		src: src,
	}
}

func (p *ParserResult) Doc() ast.Node {
	return p.doc
}

func (p *ParserResult) TableOfContents() markdown.TocFragments {
	return p.toc
}

func (p *ParserResult) Headers() []markdown.Header {
	var headers []markdown.Header

	// 遍历 AST 树，查找所有标题节点
	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if heading, ok := child.(*ast.Heading); ok {
				text := extractTextFromNode(heading, p.src)
				headers = append(headers, &HeaderNode{
					text:  text,
					level: heading.Level,

					node: heading,
					src:  p.src,
				})
			}
			walk(child) // 递归子节点
		}
	}

	walk(p.doc)
	return headers
}

func extractTextFromNode(node ast.Node, src []byte) string {
	var text string
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if segment, ok := child.(*ast.Text); ok {
			text += string(segment.Segment.Value(src))
		}
	}
	return text
}

func extractAllTextFromNode(node ast.Node, source []byte) string {
	var textBuilder strings.Builder

	// 定义 Walker 函数
	walker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Text:
			textBuilder.Write(n.Segment.Value(source)) // 提取普通文本

		case *ast.AutoLink:
			textBuilder.Write(n.Label(source)) // 提取自动链接（如邮箱、URL）

		case *ast.Link:
			textBuilder.Write(n.Text(source)) // 提取显式链接
		}

		return ast.WalkContinue, nil
	}

	// 使用 Walk 遍历整个节点树
	_ = ast.Walk(node, walker)

	return textBuilder.String()
}
