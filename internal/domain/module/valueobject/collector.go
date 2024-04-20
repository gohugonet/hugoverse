package valueobject

import (
	"strings"
)

type Collector struct {
	*GoClient

	// Pick the first and prevent circular loops.
	seen map[string]bool

	// Set if a Go modules enabled project.
	goModules GoModules

	// Ordered list of collected modules, including Go Modules and theme
	// components stored below /themes.
	modules Modules
}

func NewCollector(c *GoClient) *Collector {
	return &Collector{
		GoClient:  c,
		goModules: GoModules{},
		modules:   Modules{},
		seen:      make(map[string]bool),
	}
}

func (c *Collector) CollectGoModules() error {
	modules, err := c.listGoMods()
	if err != nil {
		return err
	}
	c.goModules = modules
	return nil
}

func (c *Collector) GetMain() *GoModule {
	for _, m := range c.goModules {
		if m.Main {
			return m
		}
	}

	return nil
}

func (c *Collector) GetGoModule(path string) *GoModule {
	if c.goModules == nil {
		return nil
	}

	for _, m := range c.goModules {
		if strings.EqualFold(path, m.Path) {
			return m
		}
	}
	return nil
}

func (c *Collector) IsSeen(path string) bool {
	key := pathKey(path)
	if c.seen[key] {
		return true
	}
	c.seen[key] = true
	return false
}
