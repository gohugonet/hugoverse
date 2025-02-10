package html

import (
	"github.com/mdfriday/hugoverse/pkg/lexer"
	"strings"
)

func readTagName(input string) (string, int) {
	pos := 0
	for {
		c, s := lexer.NextChar(input[pos:])
		//todo error check
		switch c {
		case ' ':
			panic("attribute not support yet")
		case '>':
			return input[:pos], pos
		}
		pos += s
	}
}

func readComment(input string) (string, int) {
	pos := strings.Index(input, "-->")
	if pos > 0 {
		return input[:pos], pos
	}
	return "", -1
}
