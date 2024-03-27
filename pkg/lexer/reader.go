package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const EOF = -1

func NextChar(input string) (rune, int) {
	if len(input) == 0 {
		return EOF, 0
	}

	return utf8.DecodeRuneInString(input)
}

// IsAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func IsAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func TrimLeftSpace(input string) string {
	return strings.TrimLeft(input, string('\n'))
}
