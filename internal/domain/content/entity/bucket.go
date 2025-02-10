package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/content"
)

func (c *Content) AllContentTypes() map[string]content.Creator {
	return c.UserTypes
}

func (c *Content) AllAdminTypes() map[string]content.Creator {
	return c.AdminTypes
}

func (c *Content) AllTypes() map[string]content.Creator {
	types := make(map[string]content.Creator)

	for k, v := range c.UserTypes {
		types[k] = v
	}

	for k, v := range c.AdminTypes {
		types[k] = v
	}

	return types
}

func (c *Content) GetContentCreator(name string) (content.Creator, bool) {
	t, ok := c.UserTypes[name]
	if ok {
		return t, ok
	}

	t, ok = c.AdminTypes[name]
	return t, ok
}

func (c *Content) AllContentTypeNames() []string {
	keys := make([]string, 0, len(c.UserTypes))
	for k := range c.UserTypes {
		keys = append(keys, k)
	}
	return keys
}

func (c *Content) AllAdminTypeNames() []string {
	keys := make([]string, 0, len(c.AdminTypes))
	for k := range c.AdminTypes {
		keys = append(keys, k)
	}
	return keys
}

func (c *Content) IsAdminType(name string) bool {
	_, ok := c.AdminTypes[name]
	return ok
}
