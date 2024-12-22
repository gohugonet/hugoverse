package valueobject

import "path"

type ComponentPath struct {
	Component string
	Path      string
	Lang      string
}

func (c ComponentPath) ComponentPathJoined() string {
	return path.Join(c.Component, c.Path)
}
func (c ComponentPath) GetComponent() string { return c.Component }
func (c ComponentPath) GetPath() string      { return c.Path }
func (c ComponentPath) GetLang() string      { return c.Lang }

type ReverseLookupProvider interface {
	ReverseLookup(filename string) ([]ComponentPath, error)
	ReverseLookupComponent(component, filename string) ([]ComponentPath, error)
}
