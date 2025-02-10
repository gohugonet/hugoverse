package html

import (
	"github.com/mdfriday/hugoverse/pkg/fsm"
	"github.com/mdfriday/hugoverse/pkg/lexer"
	"strings"
)

const (
	textState     = "text"
	startTagState = "startTag"
	endTagState   = "endTag"
	commentState  = "comment"
	eofState      = "eof"
)

type state struct{}

func (s *state) Init() fsm.State {
	return textState
}
func (s *state) EoF() fsm.State {
	return eofState
}
func (s *state) Mapping() map[fsm.State]lexer.StateHandler {
	return map[fsm.State]lexer.StateHandler{
		textState: func(lex lexer.Lexer) fsm.StateHandler {
			return func(event fsm.Event) (fsm.State, fsm.Data) {
				input := event.Data().Raw().(string)
				input = lexer.TrimLeftSpace(input)

				var emitTextToken = func(start int, end int) {
					if end > start {
						lex.Emit(&Token{
							BaseToken: lexer.BaseToken{
								Typ: TokenText,
								Val: input[start:end],
							},
							Start: "",
							End:   "",
						})
					}
				}

				pos := 0
				for {
					if pos = strings.Index(input[pos:], string(start)); pos >= 0 {
						c := input[pos+1]

						switch {
						case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
							emitTextToken(0, pos)
							return startTagState, &fsm.BaseData{
								Err:     nil,
								RawData: input[pos:],
							}
						case c == '/':
							emitTextToken(0, pos)
							return endTagState, &fsm.BaseData{
								Err:     nil,
								RawData: input[pos:],
							}
						case c == '!':
							emitTextToken(0, pos)
							return commentState, &fsm.BaseData{
								Err:     nil,
								RawData: input[pos:],
							}
						default:
							// not start tag
							pos++
							continue
						}
					} else {
						if len(input) > 0 {
							emitTextToken(pos, len(input))
						}
						break
					}
				}
				lex.Emit(&lexer.BaseToken{Typ: TokenEOF, Val: ""})

				return eofState, &fsm.BaseData{Err: nil, RawData: ""}
			}
		},
		startTagState: func(lex lexer.Lexer) fsm.StateHandler {
			return func(event fsm.Event) (fsm.State, fsm.Data) {
				input := event.Data().Raw().(string)

				if string(input[0]) != string(start) {
					panic("input must start with '<'")
				}
				pos := 1
				name, s := readTagName(input[pos:])
				pos += s
				for {
					c, s := lexer.NextChar(input[pos:])
					if string(c) == string(end) {
						pos += s
						break
					}
					panic("attributes not supported yet")
				}
				lex.Emit(&Token{
					BaseToken: lexer.BaseToken{Typ: TokenStartTag, Val: name},
					Start:     "<",
					End:       ">",
				})

				return textState, &fsm.BaseData{Err: nil, RawData: input[pos:]}
			}
		},
		endTagState: func(lex lexer.Lexer) fsm.StateHandler {
			return func(event fsm.Event) (fsm.State, fsm.Data) {
				input := event.Data().Raw().(string)

				if input[:2] != "</" {
					panic("input must start with '</'")
				}
				pos := len("</")
				name, s := readTagName(input[pos:])
				pos += s
				for {
					c, s := lexer.NextChar(input[pos:])
					if string(c) == string(end) {
						pos += s
						break
					}
					panic("no attributes in end tag")
				}
				lex.Emit(&Token{
					BaseToken: lexer.BaseToken{Typ: TokenEndTag, Val: name},
					Start:     "</",
					End:       ">",
				})

				return textState, &fsm.BaseData{Err: nil, RawData: input[pos:]}
			}
		},
		commentState: func(lex lexer.Lexer) fsm.StateHandler {
			return func(event fsm.Event) (fsm.State, fsm.Data) {
				input := event.Data().Raw().(string)

				if input[0:4] != "<!--" {
					panic("input must start with '<!--'")
				}
				pos := len("<!--")
				c, s := readComment(input[pos:])
				if s != -1 {
					lex.Emit(&Token{
						BaseToken: lexer.BaseToken{Typ: TokenComment, Val: c},
						Start:     "<!--",
						End:       "-->",
					})
					pos += s
					pos += len("-->")
				} else {
					//todo
				}

				return textState, &fsm.BaseData{Err: nil, RawData: input[pos:]}
			}
		},
	}
}
