package action

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/fsm"
	"github.com/gohugonet/hugoverse/pkg/lexer"
)

const (
	TokenEOF lexer.TokenType = iota
	TokenLeftDelim
	TokenRightDelim
	TokenText
	TokenField
	TokenIdentifier
	TokenPipe
)

type token struct {
	typ lexer.TokenType
	val string
}

func (t *token) Type() lexer.TokenType {
	return t.typ
}
func (t *token) Value() string {
	return t.val
}

type delim string

const (
	left  delim = "{{"
	right delim = "}}"
)

type lex struct {
	input string
	left  delim
	right delim
	token chan *token
	fsm   fsm.FSM
}

func New(input string) (lexer.Lexer, error) {
	if len(input) == 0 {
		return nil, errors.New("input is empty")
	}

	f := fsm.New(textState, &data{
		err: nil,
		raw: input,
	})

	l := &lex{
		input: input,
		left:  left,
		right: right,
		token: make(chan *token),
		fsm:   f,
	}

	initFSM(l)
	go l.run()

	return l, nil
}

func (l *lex) Next() lexer.Token {
	return <-l.token
}

func (l *lex) Tokens() lexer.Tokens {
	var tokens []lexer.Token
	for {
		// lexer iterate
		token := l.Next()
		tokens = append(tokens, token)
		// reach end, analyzing done
		if token.Type() == TokenEOF {
			break
		}
	}
	return tokens
}

func (l *lex) run() {
	for {
		if l.fsm.State() == eofState {
			fmt.Println("EOF State")
			break
		}
		e := l.fsm.Process("continue")
		if e != nil {
			fmt.Println("break because of action run error")
			break
		}
	}
	close(l.token)
}

func (l *lex) Emit(t lexer.Token) {
	l.emit(t.(*token))
}

func (l *lex) emit(t *token) {
	l.token <- t
}

type data struct {
	err error
	raw string
}

func (d *data) Error() error {
	return d.err
}
func (d *data) Raw() any {
	return d.raw
}
