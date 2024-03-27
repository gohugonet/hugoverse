package parser

import (
	"github.com/gohugonet/hugoverse/pkg/lexer"
)

func newCommand(tokens lexer.Tokens) (*commandNode, error) {
	tp := &termParser{}
	cmd := &commandNode{treeNode: &treeNode{}}
	for _, t := range tokens {
		n, _, err := tp.Parse(t)
		if err != nil {
			return nil, err
		}
		cmd.AppendChild(n)
	}

	return cmd, nil
}

type commandNode struct {
	*treeNode
	baseNode
}

func (n *commandNode) String() string {
	cs := n.Children()
	s := ""
	for _, n := range cs {
		if s != "" {
			s += " "
		}
		s += n.(Node).String()
	}
	return s
}

func (n *commandNode) Type() NodeType {
	return CommandNode
}
