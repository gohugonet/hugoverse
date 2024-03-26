package valueobject

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"html/template"
	"reflect"
	"strconv"
	"strings"
)

// DefaultConfigProvider Provider接口实现对象
type DefaultConfigProvider struct {
	Root config.Params
}

// Get 按key获取值
// 约定""键对应的是c.Root
// 嵌套获取值
func (c *DefaultConfigProvider) Get(k string) any {
	if k == "" {
		return c.Root
	}
	key, m := c.getNestedKeyAndMap(strings.ToLower(k))
	if m == nil {
		return nil
	}
	v := m[key]
	return v
}

func (c *DefaultConfigProvider) GetString(k string) string {
	v := c.Get(k)
	s, _ := ToStringE(v)
	return s
}

// getNestedKeyAndMap 支持多级查询
// 通过分隔符"."获取查询路径
func (c *DefaultConfigProvider) getNestedKeyAndMap(key string) (string, config.Params) {
	var parts []string
	parts = strings.Split(key, ".")
	current := c.Root
	for i := 0; i < len(parts)-1; i++ {
		next, found := current[parts[i]]
		if !found {
			return "", nil
		}
		var ok bool
		current, ok = next.(config.Params)
		if !ok {
			return "", nil
		}
	}
	return parts[len(parts)-1], current
}

// Set 设置键值对
// 统一key的格式为小写字母
// 如果传入的值符合Params的要求，通过root进行设置
// 如果为非Params类型，则直接赋值
func (c *DefaultConfigProvider) Set(k string, v any) {
	k = strings.ToLower(k)

	if p, ok := ToParamsAndPrepare(v); ok {
		// Set the values directly in Root.
		SetParams(c.Root, p)
	} else {
		c.Root[k] = v
	}

	return
}

// SetDefaults will set values from params if not already set.
func (c *DefaultConfigProvider) SetDefaults(params config.Params) {
	PrepareParams(params)
	for k, v := range params {
		if _, found := c.Root[k]; !found {
			c.Root[k] = v
		}
	}
}

func (c *DefaultConfigProvider) IsSet(k string) bool {
	var found bool
	key, m := c.getNestedKeyAndMap(strings.ToLower(k))
	if m != nil {
		_, found = m[key]
	}
	return found
}

// ToParamsAndPrepare converts in to Params and prepares it for use.
// If in is nil, an empty map is returned.
// See PrepareParams.
func ToParamsAndPrepare(in any) (config.Params, bool) {
	if IsNil(in) {
		return config.Params{}, true
	}
	m, err := ToStringMapE(in)
	if err != nil {
		return nil, false
	}
	PrepareParams(m)
	return m, true
}

// IsNil reports whether v is nil.
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return value.IsNil()
	}

	return false
}

// ToStringMapE converts in to map[string]interface{}.
func ToStringMapE(in any) (map[string]any, error) {
	switch vv := in.(type) {
	case config.Params:
		return vv, nil
	case map[string]any:
		var m = map[string]any{}
		for k, v := range vv {
			m[k] = v
		}
		return m, nil

	default:
		return nil, errors.New("value type not supported yet")
	}
}

// PrepareParams
// * makes all the keys lower cased
// * This will modify the map given.
// * Any nested map[string]interface{}, map[string]string
// * will be converted to Params.
func PrepareParams(m config.Params) {
	for k, v := range m {
		var retyped bool
		lKey := strings.ToLower(k)

		switch vv := v.(type) {
		case map[string]any:
			var p config.Params = v.(map[string]any)
			v = p
			PrepareParams(p)
			retyped = true
		case map[string]string:
			p := make(config.Params)
			for k, v := range vv {
				p[k] = v
			}
			v = p
			PrepareParams(p)
			retyped = true
		}

		if retyped || k != lKey {
			delete(m, k)
			m[lKey] = v
		}
	}
}

// ToStringE casts an interface to a string type.
func ToStringE(i interface{}) (string, error) {
	i = indirectToStringerOrError(i)

	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint64:
		return strconv.FormatUint(s, 10), nil
	case uint32:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(s), 10), nil
	case json.Number:
		return s.String(), nil
	case []byte:
		return string(s), nil
	case template.HTML:
		return string(s), nil
	case template.URL:
		return string(s), nil
	case template.JS:
		return string(s), nil
	case template.CSS:
		return string(s), nil
	case template.HTMLAttr:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirectToStringerOrError returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil) or an implementation of fmt.Stringer
// or error,
func indirectToStringerOrError(a interface{}) interface{} {
	if a == nil {
		return nil
	}

	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
