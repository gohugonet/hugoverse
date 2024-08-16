package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/collections"
)

var TemplateFuncsNamespaceRegistry []func() *TemplateFuncsNamespace

func AddTemplateFuncsNamespace(ns func() *TemplateFuncsNamespace) {
	TemplateFuncsNamespaceRegistry = append(TemplateFuncsNamespaceRegistry, ns)
}

func RegisterNamespaces() {
	registerCast()
	registerFmt()
	registerLang()
	registerSafe()
	registerCrypto()
	registerPath()
	registerInflect()
	registerDiagram()
	registerReflect()
	registerMath()
}

func RegisterCallbackNamespaces(cb func(ctx context.Context, name string, data any) (tmpl, res string, err error)) {
	registerPartials(cb)
}

func RegisterExtendedNamespaces(functions template.CustomizedFunctions) {
	registerCompare(functions)
	registerTransform(functions)
	registerUrls(functions)
	registerStrings(functions)
	registerResources(functions)
	registerOs(functions)
	registerSite(functions)
	registerHugo(functions)
}

func RegisterLookerNamespaces(functions template.CustomizedFunctions, looker collections.FuncLooker) {
	registerCollections(functions, looker, functions)
}
