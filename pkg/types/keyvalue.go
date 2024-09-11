package types

import (
	"fmt"
	"github.com/spf13/cast"
)

// KeyValues holds an key and a slice of values.
type KeyValues struct {
	Key    any
	Values []any
}

// KeyString returns the key as a string, an empty string if conversion fails.
func (k KeyValues) KeyString() string {
	return cast.ToString(k.Key)
}

func (k KeyValues) String() string {
	return fmt.Sprintf("%v: %v", k.Key, k.Values)
}

// KeyValueStr is a string tuple.
type KeyValueStr struct {
	Key   string
	Value string
}
