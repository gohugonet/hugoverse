package parser

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/lexer"
	"github.com/gohugonet/hugoverse/pkg/lexer/action"
)

type Document struct {
	*tree
}

func Parse(name string, text string) (*Document, error) {
	lex, err := action.New(text)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p := &parser{
		name:  name,
		lexer: lex,
		tree:  newTree(),
	}

	err = p.parse()
	if err != nil {
		return nil, err
	}

	return &Document{p.tree}, err
}

var rootParsers = map[lexer.TokenType]Parser{}

func registerRootParsers(tokenType lexer.TokenType, p Parser) {
	if _, ok := rootParsers[tokenType]; ok {
		panic("duplicated parser")
	}
	rootParsers[tokenType] = p
}

func getParser(tokenType lexer.TokenType) Parser {
	return rootParsers[tokenType]
}

type parser struct {
	name  string
	lexer lexer.Lexer
	tree  *tree
}

func (p *parser) parse() error {
	var currentParser Parser
	ps := done

	for {
		token := p.lexer.Next()

		if token.Type() == action.TokenEOF {
			break
		}

		// keep the same parser, change only after it's done
		if ps == done {
			currentParser = getParser(token.Type())
		}

		n, ps2, err := currentParser.Parse(token)
		if err != nil {
			return err
		}
		ps = ps2

		if ps == done && n != nil {
			p.tree.AppendChild(n)
		}
	}

	return nil
}
