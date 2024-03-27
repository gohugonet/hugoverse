package lexer

const (
	TokenEOF TokenType = iota
)

type BaseToken struct {
	Typ TokenType
	Val string
}

func (t *BaseToken) Type() TokenType {
	return t.Typ
}
func (t *BaseToken) Value() string {
	return t.Val
}
