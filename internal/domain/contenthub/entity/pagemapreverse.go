package entity

import "sync"

var ambiguousContentNode = &PageTreesNode{}

type contentTreeReverseIndexMap struct {
	init sync.Once
	m    map[any]*PageTreesNode
}

type contentTreeReverseIndex struct {
	initFn func(rm map[any]*PageTreesNode)
	*contentTreeReverseIndexMap
}

func (c *contentTreeReverseIndex) Reset() {
	c.contentTreeReverseIndexMap = &contentTreeReverseIndexMap{
		m: make(map[any]*PageTreesNode),
	}
}

func (c *contentTreeReverseIndex) Get(key any) *PageTreesNode {
	c.init.Do(func() {
		c.m = make(map[any]*PageTreesNode)
		c.initFn(c.contentTreeReverseIndexMap.m)
	})
	return c.m[key]
}
