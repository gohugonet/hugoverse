package resource

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/spf13/cast"
)

// New returns a new instance of the resources-namespaced template functions.
func New(resource Resource) (*Namespace, error) {
	return &Namespace{
		resourceService: resource,
	}, nil
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	resourceService Resource
}

// Get locates the filename given in Hugo's assets filesystem
// and creates a Resource object that can be used for further transformations.
func (ns *Namespace) Get(filename any) resources.Resource {
	filenamestr, err := cast.ToStringE(filename)
	if err != nil {
		panic(err)
	}

	r, err := ns.resourceService.GetResource(filenamestr)
	if err != nil {
		panic(err)
	}

	return r
}

// Copy copies r to the new targetPath in s.
func (ns *Namespace) Copy(s any, r resources.Resource) (resources.Resource, error) {
	targetPath, err := cast.ToStringE(s)
	if err != nil {
		panic(err)
	}
	return ns.resourceService.Copy(r, targetPath)
}

// GetMatch finds the first Resource matching the given pattern, or nil if none found.
//
// It looks for files in the assets file system.
//
// See Match for a more complete explanation about the rules used.
func (ns *Namespace) GetMatch(pattern any) resources.Resource {
	patternStr, err := cast.ToStringE(pattern)
	if err != nil {
		panic(err)
	}

	r, err := ns.resourceService.GetMatch(patternStr)
	if err != nil {
		panic(err)
	}

	return r
}

// Minify minifies the given Resource using the MediaType to pick the correct
// minifier.
func (ns *Namespace) Minify(r resources.Resource) (resources.Resource, error) {
	return ns.resourceService.Minify(r)
}
