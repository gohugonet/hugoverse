// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// JSONMarshaler encodes and decodes JSON and is the default used in this
// codec.
var JSONMarshaler = new(jsonMarshaler)

// New creates a new Coded with some optional options.
func New(opts ...Option) (*Codec, error) {
	c := &Codec{
		typeSep:               "|",
		marshaler:             JSONMarshaler,
		typeAdapters:          DefaultTypeAdapters,
		typeAdaptersMap:       make(map[reflect.Type]Adapter),
		typeAdaptersStringMap: make(map[string]Adapter),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return c, err
		}
	}

	for _, w := range c.typeAdapters {
		tp := w.Type()
		c.typeAdaptersMap[tp] = w
		c.typeAdaptersStringMap[tp.String()] = w
	}

	return c, nil
}

// Option configures the Codec.
type Option func(c *Codec) error

// WithTypeSep sets the separator to use before the type information encoded in
// the key field. Default is "|".
func WithTypeSep(sep string) func(c *Codec) error {
	return func(c *Codec) error {
		if sep == "" {
			return errors.New("separator cannot be empty")
		}
		c.typeSep = sep
		return nil
	}
}

// WithMarshalUnmarshaler sets the MarshalUnmarshaler to use.
// Default is JSONMarshaler.
func WithMarshalUnmarshaler(marshaler MarshalUnmarshaler) func(c *Codec) error {
	return func(c *Codec) error {
		c.marshaler = marshaler
		return nil
	}
}

// WithTypeAdapters sets the type adapters to use. Note that if more than one
// adapter exists for the same type, the last one will win. This means that
// if you want to use the default adapters, but override some of them, you
// can do:
//
//   adapters := append(typedmapcodec.DefaultTypeAdapters, mycustomAdapters ...)
//   codec := typedmapcodec.New(WithTypeAdapters(adapters))
//
func WithTypeAdapters(typeAdapters []Adapter) func(c *Codec) error {
	return func(c *Codec) error {
		c.typeAdapters = typeAdapters
		return nil
	}
}

// Codec provides methods to marshal and unmarshal a Go map while preserving
// type information.
type Codec struct {
	typeSep               string
	marshaler             MarshalUnmarshaler
	typeAdapters          []Adapter
	typeAdaptersMap       map[reflect.Type]Adapter
	typeAdaptersStringMap map[string]Adapter
}

// Marshal accepts a Go map and marshals it to the configured marshaler
// anntated with type information.
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	m, err := c.toTypedMap(v)
	if err != nil {
		return nil, err
	}
	return c.marshaler.Marshal(m)
}

// Unmarshal unmarshals the given data to the given Go map, using
// any annotated type information found to preserve the type information
// stored in Marshal.
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if err := c.marshaler.Unmarshal(data, v); err != nil {
		return err
	}
	_, err := c.fromTypedMap(v)
	return err
}

func (c *Codec) newKey(key reflect.Value, a Adapter) reflect.Value {
	return reflect.ValueOf(fmt.Sprintf("%s%s%s", key, c.typeSep, a.Type()))
}

func (c *Codec) fromTypedMap(mi interface{}) (reflect.Value, error) {
	m := reflect.ValueOf(mi)
	if m.Kind() == reflect.Ptr {
		m = m.Elem()
	}

	if m.Kind() != reflect.Map {
		return reflect.Value{}, errors.New("must be a Map")
	}

	keyKind := m.Type().Key().Kind()
	if keyKind == reflect.Interface {
		// We only support string keys.
		// YAML creates map[interface {}]interface {}, so try to convert it.
		var err error
		m, err = c.toStringMap(m)
		if err != nil {
			return reflect.Value{}, err
		}
	}

	for _, key := range m.MapKeys() {

		v := indirectInterface(m.MapIndex(key))

		var (
			keyStr   = key.String()
			keyPlain string
			keyType  string
		)

		sepIdx := strings.LastIndex(keyStr, c.typeSep)

		if sepIdx != -1 {
			keyPlain = keyStr[:sepIdx]
			keyType = keyStr[sepIdx+len(c.typeSep):]
		}

		adapter, found := c.typeAdaptersStringMap[keyType]

		if !found {
			if v.Kind() == reflect.Map {
				var err error
				v, err = c.fromTypedMap(v.Interface())
				if err != nil {
					return reflect.Value{}, err
				}
				m.SetMapIndex(key, v)
			}
			continue
		}

		switch v.Kind() {
		case reflect.Map:
			mm := reflect.MakeMap(reflect.MapOf(stringType, adapter.Type()))
			for _, key := range v.MapKeys() {
				vv := indirectInterface(v.MapIndex(key))
				nv, err := adapter.FromString(vv.String())
				if err != nil {
					return reflect.Value{}, err
				}
				mm.SetMapIndex(indirectInterface(key), reflect.ValueOf(nv))
			}
			m.SetMapIndex(reflect.ValueOf(keyPlain), mm)
		case reflect.Slice:
			slice := reflect.MakeSlice(reflect.SliceOf(adapter.Type()), v.Len(), v.Cap())
			for i := 0; i < v.Len(); i++ {
				vv := indirectInterface(v.Index(i))
				nv, err := adapter.FromString(vv.String())
				if err != nil {
					return reflect.Value{}, err
				}
				slice.Index(i).Set(reflect.ValueOf(nv))
			}

			m.SetMapIndex(reflect.ValueOf(keyPlain), slice)
		default:
			nv, err := adapter.FromString(v.String())
			if err != nil {
				return reflect.Value{}, err
			}
			m.SetMapIndex(reflect.ValueOf(keyPlain), reflect.ValueOf(nv))
		}

		m.SetMapIndex(key, reflect.Value{})

	}

	return m, nil
}

var (
	interfaceMapType   = reflect.TypeOf(make(map[string]interface{}))
	interfaceSliceType = reflect.TypeOf([]interface{}{})
	stringType         = reflect.TypeOf("")
)

func (c *Codec) toTypedMap(mi interface{}) (interface{}, error) {

	mv := reflect.ValueOf(mi)

	if mv.Kind() != reflect.Map || mv.Type().Key().Kind() != reflect.String {
		return nil, errors.New("must provide a map with string keys")
	}

	m := reflect.MakeMap(interfaceMapType)

	for _, key := range mv.MapKeys() {
		v := indirectInterface(mv.MapIndex(key))

		switch v.Kind() {

		case reflect.Map:

			if wrapper, found := c.typeAdaptersMap[v.Type().Elem()]; found {
				mm := reflect.MakeMap(interfaceMapType)
				for _, key := range v.MapKeys() {
					mm.SetMapIndex(key, reflect.ValueOf(wrapper.Wrap(v.MapIndex(key).Interface())))
				}
				m.SetMapIndex(c.newKey(key, wrapper), mm)
			} else {
				nested, err := c.toTypedMap(v.Interface())
				if err != nil {
					return nil, err
				}
				m.SetMapIndex(key, reflect.ValueOf(nested))
			}
			continue
		case reflect.Slice:
			if adapter, found := c.typeAdaptersMap[v.Type().Elem()]; found {
				slice := reflect.MakeSlice(interfaceSliceType, v.Len(), v.Cap())
				for i := 0; i < v.Len(); i++ {
					slice.Index(i).Set(reflect.ValueOf(adapter.Wrap(v.Index(i).Interface())))
				}
				m.SetMapIndex(c.newKey(key, adapter), slice)
				continue
			}
		}

		if adapter, found := c.typeAdaptersMap[v.Type()]; found {
			m.SetMapIndex(c.newKey(key, adapter), reflect.ValueOf(adapter.Wrap(v.Interface())))
		} else {
			m.SetMapIndex(key, v)
		}
	}

	return m.Interface(), nil
}

func (c *Codec) toStringMap(mi reflect.Value) (reflect.Value, error) {
	elemType := mi.Type().Elem()
	m := reflect.MakeMap(reflect.MapOf(stringType, elemType))
	for _, key := range mi.MapKeys() {
		key = indirectInterface(key)
		if key.Kind() != reflect.String {
			return reflect.Value{}, errors.New("this library supports only string keys in maps")
		}
		vv := mi.MapIndex(key)
		m.SetMapIndex(reflect.ValueOf(key.String()), vv)
	}

	return m, nil
}

// MarshalUnmarshaler is the interface that must be implemented if you want to
// add support for more than JSON to this codec.
type MarshalUnmarshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(b []byte, v interface{}) error
}

type jsonMarshaler int

func (jsonMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonMarshaler) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

// Based on: https://github.com/golang/go/blob/178a2c42254166cffed1b25fb1d3c7a5727cada6/src/text/template/exec.go#L931
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}
