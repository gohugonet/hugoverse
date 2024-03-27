package template

import (
	"github.com/gohugonet/hugoverse/pkg/template/escaper"
	"github.com/gohugonet/hugoverse/pkg/template/executor"
	"github.com/gohugonet/hugoverse/pkg/template/parser"
	"io"
)

type Template interface {
	Name() string
	Tree() *parser.Document
}

func Parse(name string, text string) (*parser.Document, error) {
	return parser.Parse(name, text)
}

func Escape(doc *parser.Document) (*parser.Document, error) {
	return escaper.Escape(doc)
}

func Execute(t Template, w io.Writer, data any) error {
	return executor.Execute(t.Tree(), t.Name(), w, data)
}
