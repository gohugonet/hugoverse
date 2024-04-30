package resources

import (
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

	r, err := ns.resourceService.Get(filenamestr)
	if err != nil {
		panic(err)
	}

	return r
}
