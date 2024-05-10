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
