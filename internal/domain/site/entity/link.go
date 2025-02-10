package entity

import (
	"github.com/mdfriday/hugoverse/pkg/paths"
	"path"
)

func (p *Page) Permalink() string {
	if p.PageIdentity().PageLanguage() == p.langSvc.DefaultLanguage() {
		return p.BaseURL.WithPathNoTrailingSlash + paths.PathEscape(p.PageOutput.TargetFilePath())
	}

	return p.BaseURL.WithPath + paths.PathEscape(
		path.Join(p.PageOutput.TargetPrefix(), p.PageOutput.TargetFilePath()))
}

func (p *Page) RelPermalink() string {
	if p.PageIdentity().PageLanguage() == p.langSvc.DefaultLanguage() {
		return p.BaseURL.WithPathNoTrailingSlash + paths.PathEscape(p.PageOutput.TargetFilePath())
	}

	return p.BaseURL.WithPath + paths.PathEscape(
		path.Join(p.PageOutput.TargetPrefix(), p.PageOutput.TargetFilePath()))
}
