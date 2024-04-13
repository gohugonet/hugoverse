package identity

// Identity represents a thing in Hugo (a Page, a template etc.)
// Any implementation must be comparable/hashable.
type Identity interface {
	IdentifierBase() string
}

// Provider can be implemented by types that isn't itself and Identity,
// usually because they're not comparable/hashable.
type Provider interface {
	GetIdentity() Identity
}
