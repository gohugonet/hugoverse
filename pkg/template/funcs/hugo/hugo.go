package hugo

// New returns a new instance of the os-namespaced template functions.
func New(info Info) *Namespace {
	return &Namespace{
		Info: info,
	}
}

// Namespace provides template functions for the "os" namespace.
type Namespace struct {
	Info
}
