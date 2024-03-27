package parser

import (
	"errors"
	"github.com/gohugonet/hugoverse/pkg/lexer"
	"github.com/gohugonet/hugoverse/pkg/lexer/action"
)

func init() {
	p := &textParser{matchingType: action.TokenText}
	registerRootParsers(p.matchingType, p)
}

type textParser struct {
	matchingType lexer.TokenType
}

func (t *textParser) Parse(token lexer.Token) (Node, ParseState, error) {
	if token.Type() != t.matchingType {
		return nil, done, errors.New("mismatch token type")
	}
	return &textNode{
		treeNode: &treeNode{},
		baseNode: baseNode{val: token.Value()},
	}, done, nil
}

type textNode struct {
	*treeNode
	baseNode
}

func (t *textNode) String() string {
	return t.val
}

func (t *textNode) Type() NodeType {
	return TextNode
}
