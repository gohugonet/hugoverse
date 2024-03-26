package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/module"
)

type ContentHub struct {
	ThemeProvider contenthub.ThemeProvider
	Modules       module.Modules

	// ExecTemplate handling.
	TemplateProvider contenthub.ResourceProvider
}
