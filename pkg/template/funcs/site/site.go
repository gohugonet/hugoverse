package site

// New returns a new instance of the resources-namespaced template functions.
func New(svc Service) (*Namespace, error) {
	return &Namespace{
		Service: svc,
	}, nil
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	Service
}
