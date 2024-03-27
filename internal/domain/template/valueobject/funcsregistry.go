package valueobject

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// TemplateFuncsNamespaceRegistry describes a registry of functions that provide
// namespaces.
var TemplateFuncsNamespaceRegistry []func() *TemplateFuncsNamespace

// TemplateFuncsNamespace represents a template function namespace.
type TemplateFuncsNamespace struct {
	// The namespace name, "strings", "lang", etc.
	Name string

	// This is the method receiver.
	Context func(v ...any) (any, error)

	// Additional info, aliases and examples, per method name.
	MethodMappings map[string]TemplateFuncMethodMapping
}

// TemplateFuncMethodMapping represents a mapping of functions to methods for a
// given namespace.
type TemplateFuncMethodMapping struct {
	Method any

	// Any template funcs aliases. This is mainly motivated by keeping
	// backwards compatibility, but some new template funcs may also make
	// sense to give short and snappy aliases.
	// Note that these aliases are global and will be merged, so the last
	// key will win.
	Aliases []string

	// A slice of input/expected examples.
	// We keep it a the namespace level for now, but may find a way to keep track
	// of the single template func, for documentation purposes.
	// Some of these, hopefully just a few, may depend on some test data to run.
	Examples [][2]string
}

// AddMethodMapping adds a method to a template function namespace.
func (t *TemplateFuncsNamespace) AddMethodMapping(m any, aliases []string, examples [][2]string) {
	if t.MethodMappings == nil {
		t.MethodMappings = make(map[string]TemplateFuncMethodMapping)
	}

	name := methodToName(m)

	// sanity check
	for _, e := range examples {
		if e[0] == "" {
			panic(t.Name + ": Empty example for " + name)
		}
	}
	for _, a := range aliases {
		if a == "" {
			panic(t.Name + ": Empty alias for " + name)
		}
	}

	t.MethodMappings[name] = TemplateFuncMethodMapping{
		Method:   m,
		Aliases:  aliases,
		Examples: examples,
	}
}

func methodToName(m any) string {
	name := runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
	name = filepath.Ext(name)
	name = strings.TrimPrefix(name, ".")
	name = strings.TrimSuffix(name, "-fm")
	return name
}

// AddTemplateFuncsNamespace adds a given function to a registry.
func AddTemplateFuncsNamespace(ns func() *TemplateFuncsNamespace) {
	TemplateFuncsNamespaceRegistry = append(TemplateFuncsNamespaceRegistry, ns)
}

func SetupRegistry() {
	f := func() *TemplateFuncsNamespace {
		ctx := &Compare{}
		ns := &TemplateFuncsNamespace{
			Name:    "compare",
			Context: func(args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Eq,
			[]string{"eq"},
			[][2]string{
				{`{{ if eq .Section "blog" }}current-section{{ end }}`, `current-section`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
