package search

import (
	"fmt"
	"strings"
)

type index struct {
	ns string
	id string
}

func (i *index) String() string {
	return fmt.Sprintf("%s:%s", i.ns, i.id)
}

func (i *index) Namespace() string {
	return i.ns
}

func (i *index) ID() string {
	return i.id
}

func CreateIndex(target string) *index {
	t := strings.Split(target, ":")
	return &index{
		ns: t[0],
		id: t[1],
	}
}

type Index interface {
	Namespace() string
	ID() string
}
