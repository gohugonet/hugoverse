package entity

import "fmt"

type Data map[string]any

func (d Data) Pages() Pages {
	v, found := d["pages"]
	if !found {
		return nil
	}

	switch vv := v.(type) {
	case []*Page:
		return vv
	case func() []*Page:
		return vv()
	default:
		panic(fmt.Sprintf("%T is not Pages", v))
	}
}
