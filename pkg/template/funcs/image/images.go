package image

// New returns a new instance of the resources-namespaced template functions.
func New(img Image) (*Namespace, error) {
	return &Namespace{
		Image: img,
	}, nil
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	Image
}
