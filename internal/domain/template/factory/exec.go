package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/entity"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"reflect"
)

func New(fs template.Fs) (template.Template, error) {
	exec, funcs := newExecutor()
	funcMap := make(map[string]any)
	for k, v := range funcs {
		funcMap[k] = v.Interface()
	}

	t := &entity.Template{
		Executor: &entity.Executor{
			Executor: exec,
		},
		Lookup: &entity.Lookup{
			Baseof:      make(map[string]valueobject.Info),
			NeedsBaseof: make(map[string]valueobject.Info),
		},
		Ast: &entity.AstTransformer{
			TransformNotFound: make(map[string]*valueobject.State),
		},

		Main: newNamespace(funcMap),
		Fs:   fs,
	}

	if err := t.LoadTemplates(); err != nil {
		return nil, err
	}

	if err := t.PostTransform(); err != nil {
		return nil, err
	}

	if err := t.Main.MarkReady(); err != nil {
		return nil, err
	}
	t.Lookup.Main = t.Main

	return t, nil
}

func newExecutor() (texttemplate.Executor, map[string]reflect.Value) {
	funcs := createFuncMap()
	funcsv := make(map[string]reflect.Value)

	for k, v := range funcs {
		vv := reflect.ValueOf(v)
		funcsv[k] = vv
	}

	// Duplicate Go's internal funcs here for faster lookups.
	for k, v := range htmltemplate.GoFuncs {
		if _, exists := funcsv[k]; !exists {
			vv, ok := v.(reflect.Value)
			if !ok {
				vv = reflect.ValueOf(v)
			}
			funcsv[k] = vv
		}
	}

	for k, v := range texttemplate.GoFuncs {
		if _, exists := funcsv[k]; !exists {
			funcsv[k] = v
		}
	}

	cb := &entity.GoTemplateCallback{
		Funcs: funcsv,
	}

	return texttemplate.NewExecuter(cb), funcsv
}

func createFuncMap() map[string]any {
	valueobject.RegisterNamespaces()

	funcMap := htmltemplate.FuncMap{}

	// Merge the namespace funcs
	for _, nsf := range valueobject.TemplateFuncsNamespaceRegistry {
		ns := nsf()
		if _, exists := funcMap[ns.Name]; exists {
			panic(ns.Name + " is a duplicate template func")
		}
		funcMap[ns.Name] = ns.Context
		for _, mm := range ns.MethodMappings {
			for _, alias := range mm.Aliases {
				if _, exists := funcMap[alias]; exists {
					panic(alias + " is a duplicate template func")
				}
				funcMap[alias] = mm.Method
			}
		}
	}

	return funcMap
}

func newNamespace(funcs map[string]any) *entity.Namespace {
	return &entity.Namespace{
		PrototypeHTML: htmltemplate.New("").Funcs(funcs),
		PrototypeText: texttemplate.New("").Funcs(funcs),
		StateMap: &valueobject.StateMap{
			Templates: make(map[string]*valueobject.State),
		},
	}
}
