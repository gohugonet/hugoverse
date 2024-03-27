package entity

import (
	"bytes"
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	pkgTmpl "github.com/gohugonet/hugoverse/pkg/template"
	"io"
	"reflect"
)

type TemplateExec struct {
	Executor template.Executor
	Funcs    map[string]reflect.Value

	*TemplateHandler
}

func (t *TemplateExec) ExecuteWithContext(ctx context.Context, tmpl template.Template, wr io.Writer, data any) error {
	return t.Executor.ExecuteWithContext(ctx, tmpl, wr, data)
}

type Executor struct {
	Helper template.ExecHelper
}

// ExecuteWithContext Note: The context is currently not fully implemeted in Hugo. This is a work in progress.
func (t *Executor) ExecuteWithContext(ctx context.Context, p template.Preparer, wr io.Writer, data any) error {
	execTmpl, err := p.Prepare()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = pkgTmpl.Execute(execTmpl, buf, data)
	if err != nil {
		return err
	}

	if _, err := wr.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

type ExecHelper struct {
	Funcs map[string]reflect.Value
}

var (
	zero             reflect.Value
	contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func (t *ExecHelper) GetFunc(ctx context.Context, tmpl template.Preparer, name string) (fn reflect.Value, firstArg reflect.Value, found bool) {
	if fn, found := t.Funcs[name]; found {
		if fn.Type().NumIn() > 0 {
			first := fn.Type().In(0)
			if first.Implements(contextInterface) {
				// TODO(bep) check if we can void this conversion every time -- and if that matters.
				// The first argument may be context.Context. This is never provided by the end user, but it's used to pass down
				// contextual information, e.g. the top level data context (e.g. Page).
				return fn, reflect.ValueOf(ctx), true
			}
		}

		return fn, zero, true
	}
	return zero, zero, false
}
