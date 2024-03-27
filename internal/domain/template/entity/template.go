package entity

import "github.com/gohugonet/hugoverse/pkg/template/parser"

type ExecTemplate struct {
	name string
	doc  *parser.Document
}

func (et *ExecTemplate) Name() string {
	return et.name
}

func (et *ExecTemplate) Tree() *parser.Document {
	return et.doc
}
