package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/glob"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/cast"
	"strings"
)

type PageResources []resources.Resource

// ByType returns resources of a given resource type (e.g. "image").
func (r PageResources) ByType(typ any) PageResources {
	tpstr, err := cast.ToStringE(typ)
	if err != nil {
		panic(err)
	}
	var filtered PageResources

	for _, resource := range r {
		rt := resource.ResourceType()
		if rt == tpstr || strings.HasPrefix(rt, tpstr) {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

// GetMatch finds the first Resource matching the given pattern, or nil if none found.
// See Match for a more complete explanation about the rules used.
func (r PageResources) GetMatch(pattern any) resources.Resource {
	patternstr, err := cast.ToStringE(pattern)
	if err != nil {
		panic(err)
	}

	g, err := glob.GetGlob(paths.AddLeadingSlash(patternstr))
	if err != nil {
		panic(err)
	}

	for _, resource := range r {
		if g.Match(paths.AddLeadingSlash(resource.Name())) {
			return resource
		}
	}

	// Finally, check the normalized name.
	for _, resource := range r {
		if nop, ok := resource.(resources.NameNormalizedProvider); ok {
			if g.Match(paths.AddLeadingSlash(nop.NameNormalized())) {
				return resource
			}
		}
	}

	return nil
}

// Get locates the name given in Resources.
// The search is case insensitive.
func (r PageResources) Get(name any) resources.Resource {
	if r == nil {
		return nil
	}
	namestr, err := cast.ToStringE(name)
	if err != nil {
		panic(err)
	}

	namestr = paths.AddLeadingSlash(namestr)

	// First check the Name.
	// Note that this can be modified by the user in the front matter,
	// also, it does not contain any language code.
	for _, resource := range r {
		if strings.EqualFold(namestr, paths.AddLeadingSlash(resource.Name())) {
			return resource
		}
	}

	// Finally, check the normalized name.
	for _, resource := range r {
		if nop, ok := resource.(resources.NameNormalizedProvider); ok {
			if strings.EqualFold(namestr, paths.AddLeadingSlash(nop.NameNormalized())) {
				return resource
			}
		}
	}

	return nil
}
