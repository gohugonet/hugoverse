package hexec

import "fmt"

// AccessDeniedError represents a security policy conflict.
type AccessDeniedError struct {
	Path     string
	Name     string
	Policies string
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied: %q is not whitelisted in policy %q; "+
		"the current security configuration is:\n\n%s\n\n", e.Name, e.Path, e.Policies)
}
