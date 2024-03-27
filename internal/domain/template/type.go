package template

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/parser"
	"io"
	"reflect"
)

// Template is the common interface between text/template and html/template.
type Template interface {
	Name() string
	Preparer
}

type ExecTemplate interface {
	Name() string
	Tree() *parser.Document
}

// Handler finds and executes templates.
type Handler interface {
	Finder
	ExecuteWithContext(ctx context.Context, t Template, wr io.Writer, data any) error
	LookupLayout(layoutNames []string) (Template, bool, error)
}

// Finder finds templates.
type Finder interface {
	Lookup
}

type Lookup interface {
	Lookup(name string) (Template, bool)
}

// Manager manages the collection of templates.
type Manager interface {
	Handler
	FuncGetter
	AddTemplate(name, tpl string) error
}

// FuncGetter allows to find a template func by name.
type FuncGetter interface {
	GetFunc(name string) (reflect.Value, bool)
}

// FuncMap is the type of the map defining the mapping from names to
// functions. Each function must have either a single return value, or two
// return values of which the second has type error. In that case, if the
// second (error) argument evaluates to non-nil during execution, execution
// terminates and Execute returns that error. FuncMap has the same base type
// as FuncMap in "text/template", copied here so clients need not import
// "text/template".
type FuncMap map[string]any

// Executor executes a given template.
type Executor interface {
	ExecuteWithContext(ctx context.Context, p Preparer, wr io.Writer, data any) error
}

// Preparer prepares the template before execution.
type Preparer interface {
	Prepare() (ExecTemplate, error)
}

// ExecHelper allows some custom eval hooks.
type ExecHelper interface {
	GetFunc(ctx context.Context, tmpl Preparer, name string) (reflect.Value, reflect.Value, bool)
}
