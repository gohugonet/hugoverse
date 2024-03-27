package executor

import (
	"github.com/gohugonet/hugoverse/pkg/template/escaper"
	"reflect"
)

type receiver struct {
	Data reflect.Value
	*escaper.Html
}

func newReceiver(data any) *receiver {
	value, ok := data.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(data)
	}

	return &receiver{Data: value}
}

var zero reflect.Value

func (r *receiver) data() reflect.Value {
	if !r.Data.IsValid() {
		return zero
	}

	rc, isNil := indirect(r.Data)
	if rc.Kind() == reflect.Interface && isNil {
		return zero
	}

	ptr := rc
	if ptr.Kind() != reflect.Interface && ptr.Kind() != reflect.Pointer && ptr.CanAddr() {
		ptr = ptr.Addr()
	}

	return ptr
}

func (r *receiver) ptr() reflect.Value {
	return reflect.ValueOf(r)
}

func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
	}
	return v, false
}
