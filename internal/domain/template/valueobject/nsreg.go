package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/collections"
)

var TemplateFuncsNamespaceRegistry []func() *TemplateFuncsNamespace

func ResetTemplateFuncsNamespaceRegistry() {
	TemplateFuncsNamespaceRegistry = nil
}

func AddTemplateFuncsNamespace(ns func() *TemplateFuncsNamespace) {
	TemplateFuncsNamespaceRegistry = append(TemplateFuncsNamespaceRegistry, ns)
}

func RegisterNamespaces() {
	registerCast()
	registerFmt()
	registerSafe()
	registerCrypto()
	registerPath()
	registerInflect()
	registerDiagram()
	registerReflect()
	registerMath()
	registerEncoding()
	registerTime()
}

func RegisterCallbackNamespaces(cb func(ctx context.Context, name string, data any) (tmpl string, res any, err error)) {
	registerPartials(cb)
}

func RegisterExtendedNamespaces(functions template.CustomizedFunctions) {
	registerLang(functions)
	registerCompare(functions)
	registerTransform(functions)
	registerUrls(functions, functions)
	registerStrings(functions)
	registerResources(functions)
	registerImages(functions)
	registerCss(functions)
	registerJs(functions)
	registerOs(functions)
	registerSite(functions)
	registerHugo(functions)
}

func RegisterLookerNamespaces(functions template.CustomizedFunctions, looker collections.FuncLooker) {
	registerCollections(functions, looker, functions)
}
