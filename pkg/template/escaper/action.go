package escaper

import (
	"github.com/gohugonet/hugoverse/pkg/template/parser"
)

func escapeActionNode(c context, n parser.Node) (context, error) {
	c = nudge(c)

	var escFuncName string
	switch c.state {
	case stateText:
		escFuncName = "EscapeHtml"
	}
	err := parser.InsertCommand(n, escFuncName)
	if err != nil {
		return context{}, err
	}

	return c, nil
}

// nudge is for state transition
// we can handle those relationships
// inside lexer
func nudge(c context) context {
	switch c.state {
	case stateTag:
		// In `<foo {{.}}`, the action should emit an attribute.
		c.state = stateAttrName
	}
	return c
}
