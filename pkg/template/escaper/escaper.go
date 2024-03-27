package escaper

import (
	"github.com/gohugonet/hugoverse/pkg/template/parser"
)

func Escape(doc *parser.Document) (*parser.Document, error) {
	var escErr error
	c := context{state: stateText}
	doc.Walk(func(n parser.Node, ws parser.WalkState) parser.WalkStatus {
		if ws == parser.WalkIn {
			switch n.Type() {
			case parser.TextNode:
				c, escErr = escapeTextNode(c, n)
			case parser.ActionNode:
				c, escErr = escapeActionNode(c, n)
			}

			if escErr != nil {
				return parser.WalkStop
			}
		}
		return parser.WalkContinue
	})

	if escErr != nil {
		return &parser.Document{}, escErr
	}

	return doc, nil
}
