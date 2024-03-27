package parser

import (
	"errors"
	"github.com/gohugonet/hugoverse/pkg/lexer"
	"github.com/gohugonet/hugoverse/pkg/lexer/action"
	"sort"
)

func init() {
	p := &actionParser{
		matchingTypeLeft:  action.TokenLeftDelim,
		matchingTypeRight: action.TokenRightDelim,
		currentMatching:   action.TokenEOF,
	}
	registerRootParsers(p.matchingTypeLeft, p)
	registerRootParsers(p.matchingTypeRight, p)
}

type actionParser struct {
	matchingTypeLeft  lexer.TokenType
	matchingTypeRight lexer.TokenType
	currentMatching   lexer.TokenType
	tokens            lexer.Tokens
}

func (p *actionParser) Parse(token lexer.Token) (Node, ParseState, error) {
	if p.currentMatching == action.TokenEOF &&
		(token.Type() != p.matchingTypeLeft && token.Type() != p.matchingTypeRight) {
		return nil, done, errors.New("mismatch token type of actionParser")
	}

	if token.Type() == p.matchingTypeLeft {
		p.currentMatching = p.matchingTypeLeft
		return nil, open, nil
	}
	if token.Type() == p.matchingTypeRight {
		an, err := newAction(p.tokens)
		if err != nil {
			return nil, done, err
		}
		return an, done, nil
	}

	p.tokens = append(p.tokens, token)
	return nil, open, nil
}

func InsertCommand(an Node, funcName string) error {
	_, ok := an.(*actionNode)
	if !ok {
		return errors.New("action node should be the parent")
	}

	cmd, err := newCommand(lexer.Tokens{&lexer.BaseToken{
		Typ: action.TokenIdentifier,
		Val: funcName,
	}})
	if err != nil {
		return err
	}

	an.AppendChild(cmd)
	return nil
}

func newAction(tokens lexer.Tokens) (*actionNode, error) {
	a := &actionNode{
		treeNode: &treeNode{},
		pipeline: pipeline{
			separator: action.TokenPipe,
		},
	}

	var seps = []int{0, len(tokens)}
	for i, t := range tokens {
		if t.Type() == a.pipeline.separator {
			seps = append(seps, i)
		}
	}
	sort.Ints(seps[:])
	for i := 1; i < len(seps); i++ {
		c, err := newCommand(tokens[i-1 : i])
		if err != nil {
			return nil, err
		}
		a.AppendChild(c)
	}

	return a, nil
}

type pipeline struct {
	separator lexer.TokenType
}

type actionNode struct {
	*treeNode
	pipeline
	baseNode
}

func (n *actionNode) String() string {
	cs := n.Children()
	s := ""
	for _, cmd := range cs {
		if s != "" {
			s += " | "
		}
		s += cmd.(*commandNode).String()
	}
	return s
}

func (n *actionNode) Type() NodeType {
	return ActionNode
}
