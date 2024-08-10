package entity

import (
	"github.com/gohugonet/hugoverse/pkg/constants"
	"sync"
)

type resourceTransformations struct {
	transformationsInit sync.Once
	transformationsErr  error
	transformations     []ResourceTransformation
}

// hasTransformationPermalinkHash reports whether any of the transformations
// in the chain creates a permalink that's based on the content, e.g. fingerprint.
func (r *resourceTransformations) hasTransformationPermalinkHash() bool {
	for _, t := range r.transformations {
		if constants.IsResourceTransformationPermalinkHash(t.Key().Name) {
			return true
		}
	}
	return false
}
