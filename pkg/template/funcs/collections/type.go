package collections

import (
	"golang.org/x/text/collate"
	"reflect"
)

type FuncLooker interface {
	GetFunc(name string) (reflect.Value, bool)
}

type Language interface {
	Collator() *collate.Collator
}
