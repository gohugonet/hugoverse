package site

// New returns a new instance of the resources-namespaced template functions.
func New(svc Service) *Namespace {
	return &Namespace{
		Service: svc,
	}
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	Service
}
