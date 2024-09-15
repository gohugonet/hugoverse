package entity

import (
	"github.com/gohugonet/hugoverse/pkg/paths"
)

func (p *Page) Permalink() string {
	return p.BaseURL.WithPathNoTrailingSlash + paths.PathEscape(p.PageOutput.TargetFilePath())
}

func (p *Page) RelPermalink() string {
	return p.BaseURL.BasePathNoTrailingSlash + paths.PathEscape(p.PageOutput.TargetFilePath())
}
