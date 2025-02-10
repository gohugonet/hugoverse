package html

import "github.com/mdfriday/hugoverse/pkg/lexer"

type Token struct {
	lexer.BaseToken
	Start lexer.Delim
	End   lexer.Delim
}
