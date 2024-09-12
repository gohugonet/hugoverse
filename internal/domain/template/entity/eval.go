package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/hreflect"
	"github.com/gohugonet/hugoverse/pkg/maps"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"reflect"
	"strings"
)

var (
	zero reflect.Value
)

type GoTemplateCallback struct {
	Funcs map[string]reflect.Value
}

func (t *GoTemplateCallback) BeforeExec(ctx context.Context, tmpl texttemplate.Preparer) {
}

func (t *GoTemplateCallback) GetFunc(ctx context.Context, tmpl texttemplate.Preparer, name string) (fn reflect.Value, firstArg reflect.Value, found bool) {
	if fn, found := t.Funcs[name]; found {
		if fn.Type().NumIn() > 0 {
			first := fn.Type().In(0)
			if hreflect.IsContextType(first) {
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

func (t *GoTemplateCallback) GetMapValue(ctx context.Context, tmpl texttemplate.Preparer, receiver, key reflect.Value) (reflect.Value, bool) {
	if params, ok := receiver.Interface().(maps.Params); ok {
		// Case insensitive.
		keystr := strings.ToLower(key.String())
		v, found := params[keystr]
		if !found {
			return zero, false
		}
		return reflect.ValueOf(v), true
	}

	v := receiver.MapIndex(key)

	return v, v.IsValid()
}

func (t *GoTemplateCallback) GetMethod(ctx context.Context, tmpl texttemplate.Preparer, receiver reflect.Value, name string) (method reflect.Value, firstArg reflect.Value) {
	fn := hreflect.GetMethodByName(receiver, name)
	if !fn.IsValid() {
		return zero, zero
	}

	if fn.Type().NumIn() > 0 {
		first := fn.Type().In(0)
		if hreflect.IsContextType(first) {
			// The first argument may be context.Context. This is never provided by the end user, but it's used to pass down
			// contextual information, e.g. the top level data context (e.g. Page).
			return fn, reflect.ValueOf(ctx)
		}
	}

	return fn, zero
}

func (t *GoTemplateCallback) OnCalled(ctx context.Context, tmpl texttemplate.Preparer, name string, args []reflect.Value, result reflect.Value) {
	// This switch is mostly for speed.
	switch name {
	case "Unmarshal":
	default:
		return
	}
}
