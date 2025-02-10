package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/config/valueobject"
	"github.com/mdfriday/hugoverse/pkg/output"
)

type OutputFormats struct {
	valueobject.OutputFormatsConfig
}

func (of OutputFormats) AllOutputFormats() output.Formats {
	return of.OutputFormatsConfig.Formats
}
