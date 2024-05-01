package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
)

// Copy copies r to the targetPath given.
func Copy(r resources.Resource, targetPath string) resources.Resource {
	if r.Err() != nil {
		panic(fmt.Sprintf("Resource has an .Err: %s", r.Err()))
	}
	return r.(resources.Copier).CloneTo(targetPath)
}
