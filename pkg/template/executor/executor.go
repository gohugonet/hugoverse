package executor

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/template/parser"
	"io"
)

func Execute(doc *parser.Document, name string, w io.Writer, data any) error {
	var exeErr error
	c := context{
		state: stateText,
		rcv:   newReceiver(data),
		w:     w,
		last:  missingVal,
	}

	doc.Walk(func(n parser.Node, ws parser.WalkState) parser.WalkStatus {
		if ws == parser.WalkIn {
			switch n.Type() {
			case parser.TextNode:
				c.state = stateText
			case parser.ActionNode:
				c, exeErr = evalActionNode(c, n)
			case parser.CommandNode:
				c, exeErr = evalCommandNode(c, n)
			case parser.FieldNode:
				c, exeErr = evalFieldNode(c, n)
			case parser.IdentifierNode:
				c, exeErr = evalIdentifierNode(c, n)
			}

			if exeErr != nil {
				return parser.WalkStop
			}
		} else if ws == parser.WalkOut {
			switch n.Type() {
			case parser.TextNode:
				if _, err := c.w.Write([]byte(n.String())); err != nil {
					panic(fmt.Sprintf("%s: text node write error %#v", name, err))
				}
			case parser.ActionNode:
				_, err := fmt.Fprint(c.w, c.last.Interface())
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		return parser.WalkContinue
	})

	if exeErr != nil {
		return exeErr
	}

	return nil
}
