package entity

import "errors"

const (
	typeNotRegistered = `Error:
There is no type registered for %[1]s

Add this to the file which defines %[1]s{} in the 'content' package:


	func init() {			
		item.Types["%[1]s"] = func() interface{} { return new(%[1]s) }
	}		
				

`
)

var (
	// ErrTypeNotRegistered means content type isn't registered (not found in Types map)
	ErrTypeNotRegistered = errors.New(typeNotRegistered)

	// ErrAllowHiddenItem should be used as an error to tell a caller of Hideable#Hide
	// that this type is hidden, but should be shown in a particular case, i.e.
	// if requested by a valid admin or user
	ErrAllowHiddenItem = errors.New(`Allow hidden item`)
)

type Content struct {
	Types map[string]func() interface{}
}

func (c *Content) AllContentTypeNames() []string {
	keys := make([]string, 0, len(c.Types))
	for k := range c.Types {
		keys = append(keys, k)
	}
	return keys
}
