package valueobject

import (
	"fmt"
	"strings"
)

type Index struct {
	ns string
	id string
}

func (i *Index) String() string {
	return fmt.Sprintf("%s:%s", i.ns, i.id)
}

func (i *Index) Namespace() string {
	return i.ns
}

func (i *Index) ContentType() string {
	return i.ns
}

func (i *Index) ID() string {
	return i.id
}

func CreateIndex(target string) *Index {
	t := strings.Split(target, ":")
	return &Index{
		ns: t[0],
		id: t[1],
	}
}

func NewIndex(ns, id string) *Index {
	return &Index{
		ns: ns,
		id: id,
	}
}
