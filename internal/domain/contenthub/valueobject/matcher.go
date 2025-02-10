package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/pkg/glob"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
	"strings"
)

// A PageMatcher can be used to match a Page with Glob patterns.
// Note that the pattern matching is case insensitive.
type PageMatcher struct {
	// A Glob pattern matching the content path below /content.
	// Expects Unix-styled slashes.
	// Note that this is the virtual path, so it starts at the mount root
	// with a leading "/".
	Path string

	// A Glob pattern matching the Page's Kind(s), e.g. "{home,section}"
	Kind string

	// A Glob pattern matching the Page's language, e.g. "{en,sv}".
	Lang string

	// A Glob pattern matching the Page's Environment, e.g. "{production,development}".
	Environment string
}

// Matches returns whether p matches this matcher.
func (m PageMatcher) Matches(p contenthub.Page) bool {
	if m.Kind != "" {
		g, err := glob.GetGlob(m.Kind)
		if err == nil && !g.Match(p.Kind()) {
			return false
		}
	}

	if m.Lang != "" {
		g, err := glob.GetGlob(m.Lang)
		if err == nil && !g.Match(p.PageIdentity().PageLanguage()) {
			return false
		}
	}

	if m.Path != "" {
		g, err := glob.GetGlob(m.Path)
		// TODO(bep) RelPath() vs filepath vs leading slash.
		p := strings.ToLower(filepath.ToSlash(p.Paths().Path()))
		if !(strings.HasPrefix(p, "/")) {
			p = "/" + p
		}
		if err == nil && !g.Match(p) {
			return false
		}
	}

	return true
}

type PageMatcherParamsConfig struct {
	// Apply Params to all Pages matching Target.
	Params maps.Params
	Target PageMatcher
}

func (p *PageMatcherParamsConfig) init() error {
	maps.PrepareParams(p.Params)
	return nil
}

// decodePageMatcher decodes m into v.
func decodePageMatcher(m any, v *PageMatcher) error {
	if err := mapstructure.WeakDecode(m, v); err != nil {
		return err
	}

	v.Kind = strings.ToLower(v.Kind)
	if v.Kind != "" {
		g, _ := glob.GetGlob(v.Kind)
		found := false
		for _, k := range AllKindsInPages {
			if g.Match(k) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%q did not match a valid Page Kind", v.Kind)
		}
	}

	v.Path = filepath.ToSlash(strings.ToLower(v.Path))

	return nil
}
