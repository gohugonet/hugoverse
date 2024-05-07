package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/identity"
	pio "github.com/gohugonet/hugoverse/pkg/io"
)

// ResourceTransformationKey are provided by the different transformation implementations.
// It identifies the transformation (name) and its configuration (elements).
// We combine this in a chain with the rest of the transformations
// with the target filename and a content hash of the origin to use as cache key.
type ResourceTransformationKey struct {
	Name     string
	elements []any
}

// NewResourceTransformationKey creates a new ResourceTransformationKey from the transformation
// name and elements. We will create a 64 bit FNV hash from the elements, which when combined
// with the other key elements should be unique for all practical applications.
func NewResourceTransformationKey(name string, elements ...any) ResourceTransformationKey {
	return ResourceTransformationKey{Name: name, elements: elements}
}

// Value returns the Key as a string.
// Do not change this without good reasons.
func (k ResourceTransformationKey) Value() string {
	if len(k.elements) == 0 {
		return k.Name
	}

	return k.Name + "_" + identity.HashString(k.elements...)
}

// ContentReadSeekerCloser returns a ReadSeekerCloser if possible for a given Resource.
func ContentReadSeekerCloser(r resources.Resource) (pio.ReadSeekCloser, error) {
	switch rr := r.(type) {
	case resources.ReadSeekCloserResource:
		rc, err := rr.ReadSeekCloser()
		if err != nil {
			return nil, err
		}
		return rc, nil
	default:
		return nil, fmt.Errorf("cannot transform content of Resource of type %T", r)

	}
}
