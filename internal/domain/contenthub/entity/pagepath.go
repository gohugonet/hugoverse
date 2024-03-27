package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
)

type pagePaths struct {
	outputFormats     valueobject.OutputFormats
	firstOutputFormat valueobject.OutputFormat

	targetPaths          map[string]targetPathsHolder
	targetPathDescriptor TargetPathDescriptor
}

func (l pagePaths) OutputFormats() valueobject.OutputFormats {
	return l.outputFormats
}
