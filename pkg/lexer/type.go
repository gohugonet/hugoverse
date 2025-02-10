package lexer

import "github.com/mdfriday/hugoverse/pkg/fsm"

type Delim string
type TokenType int
type Tokens []Token

type Token interface {
	Type() TokenType
	Value() string
}

type Lexer interface {
	Next() Token
	Tokens() Tokens
	Emit(t Token)
}

type StateHandler func(lex Lexer) fsm.StateHandler

type FSMState interface {
	Init() fsm.State
	EoF() fsm.State
	Mapping() map[fsm.State]StateHandler
}
