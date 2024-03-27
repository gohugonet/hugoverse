package html

import "github.com/gohugonet/hugoverse/pkg/lexer"

type Token struct {
	lexer.BaseToken
	Start lexer.Delim
	End   lexer.Delim
}
