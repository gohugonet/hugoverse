package entity

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	htemplate "html/template"
	"io"
)

func (t *Template) Execute(ctx context.Context, name string, data any) (tmpl string, res string, err error) {
	templ, found := t.Main.Lookup(name)
	if !found {
		// For legacy reasons.
		templ, found = t.Main.Lookup(name + ".html")
	}

	if !found {
		return "", "", fmt.Errorf("partial %q not found", name)
	}

	var info template.ParseInfo
	if ip, ok := templ.(template.Info); ok {
		info = ip.ParseInfo()
	} else {
		panic("not implemented template info: `ParseInfo() ParseInfo`")
	}

	var w io.Writer

	if info.Return() {
		// Wrap the context sent to the template to capture the return value.
		// Note that the template is rewritten to make sure that the dot (".")
		// and the $ variable points to Arg.
		data = &contextWrapper{
			Arg: data,
		}

		// We don't care about any template output.
		w = io.Discard
	} else {
		b := bp.GetBuffer()
		defer bp.PutBuffer(b)
		w = b
	}

	if err := t.Executor.ExecuteWithContext(ctx, templ, w, data); err != nil {
		return "", "", err
	}

	var result any

	if ctx, ok := data.(*contextWrapper); ok {
		result = ctx.Result.(string)
	} else if _, ok := templ.(*texttemplate.Template); ok {
		result = w.(fmt.Stringer).String()
	} else {
		result = string(htemplate.HTML(w.(fmt.Stringer).String()))
	}

	return templ.Name(), result.(string), nil
}

// contextWrapper makes room for a return value in a partial invocation.
type contextWrapper struct {
	Arg    any
	Result any
}

// Set sets the return value and returns an empty string.
func (c *contextWrapper) Set(in any) string {
	c.Result = in
	return ""
}
