package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/config"

// Language manages specific-language configuration.
type Language struct {
	Lang   string
	Weight int // for sort

	// If set per language, this tells Hugo that all content files without any
	// language indicator (e.g. my-page.en.md) is in this language.
	// This is usually a pathspec relative to the working dir, but it can be an
	// absolute directory reference. It is what we get.
	// For internal use.
	ContentDir string

	// Global config.
	// For internal use.
	Cfg config.Provider

	// Language specific config.
	// For internal use.
	LocalCfg config.Provider

	// Composite config.
	// For internal use.
	config.Provider
}

func (l *Language) Language() string {
	return l.Lang
}

// Languages is a sortable list of language.
type Languages []*Language

func (l Languages) Len() int { return len(l) }
func (l Languages) Less(i, j int) bool {
	wi, wj := l[i].Weight, l[j].Weight

	if wi == wj {
		return l[i].Lang < l[j].Lang
	}

	return wj == 0 || wi < wj
}

func (l Languages) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
