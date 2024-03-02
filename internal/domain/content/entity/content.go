package entity

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

func (c *Content) GetContent(name string) (func() interface{}, bool) {
	t, ok := c.Types[name]
	return t, ok
}
