package parser

import (
	"github.com/gohugonet/hugoverse/pkg/lexer"
	"github.com/gohugonet/hugoverse/pkg/lexer/action"
)

type termParser struct {
}

func (p *termParser) Parse(token lexer.Token) (Node, ParseState, error) {
	switch token.Type() {
	case action.TokenField:
		f := &fieldNode{
			treeNode: &treeNode{},
			baseNode: baseNode{val: token.Value()},
		}
		return f, done, nil
	case action.TokenIdentifier:
		i := &identifierNode{
			treeNode: &treeNode{},
			baseNode: baseNode{val: token.Value()},
		}
		return i, done, nil
	default:
		panic("not supported type token yet")
	}

	return nil, done, nil
}

type fieldNode struct {
	*treeNode
	baseNode
}

func (n *fieldNode) String() string {
	return n.val
}

func (n *fieldNode) Type() NodeType {
	return FieldNode
}

type identifierNode struct {
	*treeNode
	baseNode
}

func (i *identifierNode) String() string {
	return i.val
}

func (i *identifierNode) Type() NodeType {
	return IdentifierNode
}
