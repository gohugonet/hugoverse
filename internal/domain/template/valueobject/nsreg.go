package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/template"
)

var TemplateFuncsNamespaceRegistry []func() *TemplateFuncsNamespace

func AddTemplateFuncsNamespace(ns func() *TemplateFuncsNamespace) {
	TemplateFuncsNamespaceRegistry = append(TemplateFuncsNamespaceRegistry, ns)
}

func RegisterNamespaces() {
	registerCast()
	registerFmt()
	registerCompare()
	registerLang()
}

func RegisterCallbackNamespaces(cb func(ctx context.Context, name string, data any) (tmpl, res string, err error)) {
	registerPartials(cb)
}

func RegisterMarkdownNamespaces(mdService template.CustomizedFunctions) {
	registerTransform(mdService)
}
