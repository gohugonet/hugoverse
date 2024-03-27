package lexer

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/fsm"
)

func New(input string, sms FSMState) (Lexer, error) {
	if len(input) == 0 || sms == nil {
		return nil, errors.New("both input and sm are required")
	}

	f := fsm.New(sms.Init(), &fsm.BaseData{
		Err:     nil,
		RawData: input,
	})

	l := &lex{
		input:     input,
		token:     make(chan Token),
		fsm:       f,
		initState: sms.Init(),
		eofState:  sms.EoF(),
	}

	l.initFSM(sms.Mapping())
	go l.run()

	return l, nil
}

type lex struct {
	input     string
	token     chan Token
	fsm       fsm.FSM
	initState fsm.State
	eofState  fsm.State
}

func (l *lex) initFSM(mapping map[fsm.State]StateHandler) {
	for k, h := range mapping {
		l.fsm.Add(k, h(l))
	}
}

func (l *lex) Emit(t Token) {
	l.token <- t
}

func (l *lex) Next() Token {
	return <-l.token
}

func (l *lex) Tokens() Tokens {
	var tokens []Token
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
		if l.fsm.State() == l.eofState {
			fmt.Println("EOF State")
			break
		}
		e := l.fsm.Process("continue")
		if e != nil {
			fmt.Println("break because of run error")
			break
		}
	}
	close(l.token)
}
