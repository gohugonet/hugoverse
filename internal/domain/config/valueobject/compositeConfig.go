package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/config"

// CompositeConfig contains a read only config Base with
// a possibly writeable config Layer on top.
type CompositeConfig struct {
	Base  config.Provider
	Layer config.Provider
}

func (c *CompositeConfig) Get(key string) any {
	if c.Layer.IsSet(key) {
		return c.Layer.Get(key)
	}
	return c.Base.Get(key)
}

func (c *CompositeConfig) IsSet(key string) bool {
	if c.Layer.IsSet(key) {
		return true
	}
	return c.Base.IsSet(key)
}

func (c *CompositeConfig) GetString(key string) string {
	if c.Layer.IsSet(key) {
		return c.Layer.GetString(key)
	}
	return c.Base.GetString(key)
}

func (c *CompositeConfig) Set(key string, value any) {
	c.Layer.Set(key, value)
}

func (c *CompositeConfig) SetDefaults(params config.Params) {
	c.Layer.SetDefaults(params)
}
