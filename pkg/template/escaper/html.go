package escaper

import (
	"html/template"
	"reflect"
)

type contentType uint8

const (
	contentTypePlain contentType = iota
	contentTypeHTML
)

type Html struct {
}

func (h *Html) EscapeHtml(args ...any) string {
	s, t := stringify(args...)
	if t == contentTypeHTML {
		return s
	}
	panic("not support other content type yet")
}

// stringify converts its arguments to a string and the type of the content.
// All pointers are dereferenced, as in the text/template package.
func stringify(args ...any) (string, contentType) {
	if len(args) == 1 {
		switch s := indirect(args[0]).(type) {
		case string:
			return s, contentTypePlain
		case template.HTML:
			return string(s), contentTypeHTML
		}
	}
	return "", contentTypePlain
}

// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a any) any {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Pointer {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
