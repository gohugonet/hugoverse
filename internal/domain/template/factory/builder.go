package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/entity"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"reflect"
	"sync"
)

type builder struct {
	tmpl    *entity.Template
	funcMap map[string]any
	funcsv  map[string]reflect.Value
	cfs     template.CustomizedFunctions
}

func newBuilder() *builder {
	return &builder{
		tmpl: &entity.Template{
			Log:                 loggers.NewDefault(),
			LayoutTemplateCache: make(map[string]valueobject.LayoutCacheEntry),
		},
	}
}

func (b *builder) build() (*entity.Template, error) {
	if err := b.tmpl.LoadEmbedded(); err != nil {
		return nil, err
	}

	if err := b.tmpl.LoadTemplates(); err != nil {
		return nil, err
	}

	if err := b.tmpl.PostTransform(); err != nil {
		return nil, err
	}

	if err := b.tmpl.Parser.MarkReady(); err != nil {
		return nil, err
	}

	return b.tmpl, nil
}

func (b *builder) withFs(fs template.Fs) *builder {
	b.tmpl.Fs = fs
	return b
}

func (b *builder) withNamespace(ns *entity.Namespace) *builder {
	b.tmpl.Main = ns
	return b
}

func (b *builder) buildLookup() *builder {
	b.tmpl.Lookup = newLookup(b.funcsv)
	return b
}

func (b *builder) withCfs(cfs template.CustomizedFunctions) *builder {
	b.cfs = cfs
	return b
}

func (b *builder) buildFunctions() *builder {
	valueobject.RegisterNamespaces()
	valueobject.RegisterCallbackNamespaces(b.tmpl.Execute)
	valueobject.RegisterExtendedNamespaces(b.cfs)
	valueobject.RegisterLookerNamespaces(b.cfs, b.tmpl.Lookup)

	funcs := htmltemplate.FuncMap{}

	// Merge the namespace funcs
	for _, nsf := range valueobject.TemplateFuncsNamespaceRegistry {
		ns := nsf()
		if _, exists := funcs[ns.Name]; exists {
			continue
		}
		funcs[ns.Name] = ns.Context
		for _, mm := range ns.MethodMappings {
			for _, alias := range mm.Aliases {
				if _, exists := funcs[alias]; exists {
					continue
				}
				funcs[alias] = mm.Method
			}
		}
	}

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

	funcMap := make(map[string]any)
	for k, v := range funcsv {
		funcMap[k] = v.Interface()
	}

	b.funcsv = funcsv
	b.funcMap = funcMap
	return b
}

func (b *builder) buildParser() *builder {
	b.tmpl.Parser = &entity.Parser{
		PrototypeHTML: htmltemplate.New("").Funcs(b.funcMap),
		PrototypeText: texttemplate.New("").Funcs(b.funcMap),
		Ast: &entity.AstTransformer{
			TransformNotFound: make(map[string]*valueobject.State),
		},

		RWMutex: &sync.RWMutex{},
	}
	return b
}

func (b *builder) buildExecutor() *builder {
	cb := &entity.GoTemplateCallback{
		Funcs: b.funcsv,
	}
	b.tmpl.Executor = &entity.Executor{
		Executor: texttemplate.NewExecuter(cb),
	}
	return b
}
