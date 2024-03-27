package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/entity"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"github.com/spf13/afero"
	"reflect"
)

func NewTemplateExec(layout afero.Fs) (*entity.TemplateExec, error) {
	exec, funcs := newTemplateExecutor()
	funcMap := make(map[string]any)
	for k, v := range funcs {
		funcMap[k] = v.Interface()
	}

	h := &entity.TemplateHandler{
		Main:     newTemplateNamespace(funcMap),
		LayoutFs: layout,
	}

	if err := h.LoadTemplates(); err != nil {
		return nil, err
	}

	e := &entity.TemplateExec{
		Executor:        exec,
		Funcs:           funcs,
		TemplateHandler: h,
	}

	return e, nil
}

func newTemplateNamespace(funcs map[string]any) *entity.TemplateNamespace {
	return &entity.TemplateNamespace{
		PrototypeHTML: newHtmlTemplate("").Funcs(funcs),
		PrototypeText: newTextTemplate("").Funcs(funcs),
		TemplateStateMap: &entity.TemplateStateMap{
			Templates: make(map[string]*entity.TemplateState),
		},
	}
}

// New allocates a new HTML template with the given name.
func newHtmlTemplate(name string) *entity.HtmlTemplate {
	tmpl := &entity.HtmlTemplate{
		Text:      newTextTemplate(name),
		NameSpace: &entity.NameSpace{Set: map[string]*entity.HtmlTemplate{}},
	}
	tmpl.Set[name] = tmpl
	return tmpl
}

// New allocates a new, undefined template with the given name.
func newTextTemplate(name string) *entity.TextTemplate {
	t := &entity.TextTemplate{
		Name: name,
	}
	t = t.New(name)

	return t
}

func newTemplateExecutor() (template.Executor, map[string]reflect.Value) {
	functions := createFuncMap()
	fsv := make(map[string]reflect.Value)
	for k, v := range functions {
		vv := reflect.ValueOf(v)
		fsv[k] = vv
	}

	// Simplify
	// Duplicate Go's internal funcs here for faster lookups.
	// Build in functions

	exeHelper := &entity.ExecHelper{
		Funcs: fsv,
	}

	return &entity.Executor{Helper: exeHelper}, fsv
}

func createFuncMap() map[string]any {
	funcMap := template.FuncMap{}

	valueobject.SetupRegistry()
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
