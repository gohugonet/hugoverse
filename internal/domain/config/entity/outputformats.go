package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/output"
)

type OutputFormats struct {
	valueobject.OutputFormatsConfig
}

func (of OutputFormats) AllOutputFormats() output.Formats {
	return of.OutputFormatsConfig.Formats
}
