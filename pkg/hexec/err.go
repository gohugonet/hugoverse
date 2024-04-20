package hexec

import "fmt"

// AccessDeniedError represents a security policy conflict.
type AccessDeniedError struct {
	path     string
	name     string
	policies string
}

func (e *AccessDeniedError) Error() string {
	return fmt.Sprintf("access denied: %q is not whitelisted in policy %q; the current security configuration is:\n\n%s\n\n", e.name, e.path, e.policies)
}
