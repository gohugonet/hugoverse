package executor

import (
	"fmt"
	"reflect"
)

type missingValType struct{}

var missingVal = reflect.ValueOf(missingValType{})

func evalCall(fun, last reflect.Value) reflect.Value {
	typ := fun.Type()
	numIn := 0

	if last != missingVal {
		numIn++
	}

	// check arg count
	if typ.IsVariadic() {
		// last arg is variadic one
		if numIn < typ.NumIn()-1 {
			panic("numIn should larger than fixed args in")
		}
	} else if numIn != typ.NumIn() {
		panic("numIn should equal to fixed args in")
	}

	argv := make([]reflect.Value, numIn)
	if last != missingVal {
		// last arg type
		t := typ.In(typ.NumIn() - 1)
		if typ.IsVariadic() {
			// todo:
			// if numIn less than function fixed argc - 1
			// that means numIn last value not the variadic one
			//
			// assume last one is the variadic one
			t = t.Elem()
		}
		argv[numIn-1] = validateType(last, t)
	}

	v, err := safeCall(fun, argv)
	// If we have an error that is not nil, stop execution and return that
	// error to the caller.
	if err != nil {
		_ = fmt.Errorf("error calling %s: %w", fun.String(), err)
	}
	return unwrap(v)
}

func unwrap(v reflect.Value) reflect.Value {
	if v.Type() == reflectValueType {
		v = v.Interface().(reflect.Value)
	}
	return v
}

// safeCall runs fun.Call(args), and returns the resulting value and error, if
// any. If the call panics, the panic value is returned as an error.
func safeCall(fun reflect.Value, args []reflect.Value) (val reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	ret := fun.Call(args)
	if len(ret) == 2 && !ret[1].IsNil() {
		return ret[0], ret[1].Interface().(error)
	}
	return ret[0], nil
}

// validateType guarantees that the value is valid and assignable to the type.
func validateType(value reflect.Value, typ reflect.Type) reflect.Value {
	if !value.IsValid() {
		if typ == nil {
			// An untyped nil interface{}. Accept as a proper nil value.
			return reflect.ValueOf(nil)
		}
		if canBeNil(typ) {
			// Like above, but use the zero value of the non-nil type.
			return reflect.Zero(typ)
		}
		fmt.Printf("invalid value; expected %s\n", typ)
	}
	// get real value from reflect
	if typ == reflectValueType && value.Type() != typ {
		return reflect.ValueOf(value)
	}
	// cannot directly assign to type
	// need to be converted
	if typ != nil && !value.Type().AssignableTo(typ) {
		if value.Kind() == reflect.Interface && !value.IsNil() {
			value = value.Elem()
			if value.Type().AssignableTo(typ) {
				return value
			}
			// fallthrough
		}
		// Does one dereference or indirection work? We could do more, as we
		// do with method receivers, but that gets messy and method receivers
		// are much more constrained, so it makes more sense there than here.
		// Besides, one is almost always all you need.
		switch {
		case value.Kind() == reflect.Pointer && value.Type().Elem().AssignableTo(typ):
			value = value.Elem()
			if !value.IsValid() {
				fmt.Printf("dereference of nil pointer of type %s\n", typ)
			}
		case reflect.PointerTo(value.Type()).AssignableTo(typ) && value.CanAddr():
			value = value.Addr()
		default:
			fmt.Printf("wrong type for value; expected %s; got %s\n", typ, value.Type())
		}
	}
	return value
}

var (
	reflectValueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()
)

// canBeNil reports whether an untyped nil can be assigned to the type. See reflect.Zero.
func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	case reflect.Struct:
		return typ == reflectValueType
	}
	return false
}
