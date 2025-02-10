package html

import (
	"errors"
	"github.com/mdfriday/hugoverse/pkg/lexer"
)

const (
	TokenEOF lexer.TokenType = iota
	TokenText
	TokenStartTag
	TokenEndTag
	TokenComment
)

const (
	start lexer.Delim = "<"
	end   lexer.Delim = ">"
)

func New(input string) (lexer.Lexer, error) {
	if len(input) == 0 {
		return nil, errors.New("input is empty")
	}

	l, err := lexer.New(input, &state{})
	if err != nil {
		return nil, err
	}

	return l, nil
}
