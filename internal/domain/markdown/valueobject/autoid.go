package valueobject

import (
	"bytes"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/text"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
	"strconv"
	"unicode"
	"unicode/utf8"
)

var _ parser.IDs = (*IdFactory)(nil)

type IdFactory struct {
	idType string
	vals   map[string]struct{}
}

func NewIDFactory(idType string) *IdFactory {
	return &IdFactory{
		vals:   make(map[string]struct{}),
		idType: idType,
	}
}

func (ids *IdFactory) Generate(value []byte, kind ast.NodeKind) []byte {
	return sanitizeAnchorNameWithHook(value, ids.idType, func(buf *bytes.Buffer) {
		if buf.Len() == 0 {
			if kind == ast.KindHeading {
				buf.WriteString("heading")
			} else {
				buf.WriteString("id")
			}
		}

		if _, found := ids.vals[util.BytesToReadOnlyString(buf.Bytes())]; found {
			// Append a hyphen and a number, starting with 1.
			buf.WriteRune('-')
			pos := buf.Len()
			for i := 1; ; i++ {
				buf.WriteString(strconv.Itoa(i))
				if _, found := ids.vals[util.BytesToReadOnlyString(buf.Bytes())]; !found {
					break
				}
				buf.Truncate(pos)
			}
		}

		ids.vals[buf.String()] = struct{}{}
	})
}

func (ids *IdFactory) Put(value []byte) {
	ids.vals[util.BytesToReadOnlyString(value)] = struct{}{}
}

func sanitizeAnchorNameWithHook(b []byte, idType string, hook func(buf *bytes.Buffer)) []byte {
	buf := bp.GetBuffer()

	if idType == AutoHeadingIDTypeBlackfriday {
		panic("not implemented yet for AutoHeadingIDTypeBlackfriday")

	} else {
		asciiOnly := idType == AutoHeadingIDTypeGitHubAscii

		if asciiOnly {
			// Normalize it to preserve accents if possible.
			b = text.RemoveAccents(b)
		}

		b = bytes.TrimSpace(b)

		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)
			switch {
			case asciiOnly && size != 1:
			case r == '-' || r == ' ':
				buf.WriteRune('-')
			case isAlphaNumeric(r):
				buf.WriteRune(unicode.ToLower(r))
			default:
			}

			b = b[size:]
		}
	}

	if hook != nil {
		hook(buf)
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	bp.PutBuffer(buf)

	return result
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
