package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"strings"
)

type Cascade struct {
	params map[PageMatcher]maps.Params
}

func NewCascade(cas any) (*Cascade, error) {
	cascade, err := DecodeCascadeConfig(cas)
	if err != nil {
		return nil, err
	}
	return &Cascade{params: cascade}, nil
}

func DecodeCascadeConfig(in any) (map[PageMatcher]maps.Params, error) {
	cascade := make(map[PageMatcher]maps.Params)
	if in == nil {
		return cascade, nil
	}
	ms, err := maps.ToSliceStringMap(in)
	if err != nil {
		return nil, err
	}

	var cfgs []PageMatcherParamsConfig

	for _, m := range ms {
		m = maps.CleanConfigStringMap(m)
		c, err := mapToPageMatcherParamsConfig(m)
		if err != nil {
			return nil, err
		}
		for k := range m {
			if disallowedCascadeKeys[k] {
				return nil, fmt.Errorf("key %q not allowed in cascade config", k)
			}
		}
		cfgs = append(cfgs, c)
	}

	for _, cfg := range cfgs {
		m := cfg.Target
		CheckCascadePattern(m)
		c, found := cascade[m]
		if found {
			// Merge
			for k, v := range cfg.Params {
				if _, found := c[k]; !found {
					c[k] = v
				}
			}
		} else {
			cascade[m] = cfg.Params
		}
	}

	return cascade, nil
}

func CheckCascadePattern(m PageMatcher) {
	if isGlobWithExtension(m.Path) {
		fmt.Println("cascade-pattern-with-extension", "cascade target path %q looks like a path with an extension; since Hugo v0.123.0 this will not match anything, see  https://gohugo.io/methods/page/path/", m.Path)
	}
}

// See issue 11977.
func isGlobWithExtension(s string) bool {
	pathParts := strings.Split(s, "/")
	last := pathParts[len(pathParts)-1]
	return strings.Count(last, ".") > 0
}

var disallowedCascadeKeys = map[string]bool{
	// These define the structure of the page tree and cannot
	// currently be set in the cascade.
	"kind": true,
	"path": true,
	"lang": true,
}

func mapToPageMatcherParamsConfig(m map[string]any) (PageMatcherParamsConfig, error) {
	var pcfg PageMatcherParamsConfig
	for k, v := range m {
		switch strings.ToLower(k) {
		case "params":
			// We simplified the structure of the cascade config in Hugo 0.111.0.
			// There is a small chance that someone has used the old structure with the params keyword,
			// those values will now be moved to the top level.
			// This should be very unlikely as it would lead to constructs like .Params.params.foo,
			// and most people see params as an Hugo internal keyword.
			params := maps.ToStringMap(v)
			if pcfg.Params == nil {
				pcfg.Params = params
			} else {
				for k, v := range params {
					if _, found := pcfg.Params[k]; !found {
						pcfg.Params[k] = v
					}
				}
			}
		case "_target", "target":
			var target PageMatcher
			if err := decodePageMatcher(v, &target); err != nil {
				return pcfg, err
			}
			pcfg.Target = target
		default:
			// Legacy config.
			if pcfg.Params == nil {
				pcfg.Params = make(maps.Params)
			}
			pcfg.Params[k] = v
		}
	}
	return pcfg, pcfg.init()
}
