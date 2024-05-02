package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/constants"
	"sync"
)

// These are transformations that need special support in Hugo that may not
// be available when building the theme/site so we write the transformation
// result to disk and reuse if needed for these,
// TODO(bep) it's a little fragile having these constants redefined here.
var transformationsToCacheOnDisk = map[string]bool{
	"postcss":    true,
	"tocss":      true,
	"tocss-dart": true,
}

type resourceTransformations struct {
	transformationsInit sync.Once
	transformationsErr  error
	transformations     []valueobject.ResourceTransformation
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
