package identity

import "github.com/gohugonet/hugoverse/pkg/types"

const (
	// Anonymous is an Identity that can be used when identity doesn't matter.
	Anonymous = StringIdentity("__anonymous")

	// GenghisKhan is an Identity everyone relates to.
	GenghisKhan = StringIdentity("__genghiskhan")
)

// StringIdentity is an Identity that wraps a string.
type StringIdentity string

func (s StringIdentity) IdentifierBase() string {
	return string(s)
}

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

// Manager  is an Identity that also manages identities, typically dependencies.
type Manager interface {
	Identity
	AddIdentity(ids ...Identity)
	AddIdentityForEach(ids ...ForEeachIdentityProvider)
	GetIdentity() Identity
	Reset()
	forEeachIdentity(func(id Identity) bool) bool
}

// ForEeachIdentityProvider provides a way iterate over identities.
type ForEeachIdentityProvider interface {
	// ForEeachIdentityProvider calls cb for each Identity.
	// If cb returns true, the iteration is terminated.
	// The return value is whether the iteration was terminated.
	ForEeachIdentity(cb func(id Identity) bool) bool
}

// WalkIdentitiesShallow will not walk into a Manager's Identities.
// See WalkIdentitiesDeep.
// cb is called for every Identity found and returns whether to terminate the walk.
func WalkIdentitiesShallow(v any, cb func(level int, id Identity) bool) {
	walkIdentitiesShallow(v, 0, cb)
}

func walkIdentitiesShallow(v any, level int, cb func(level int, id Identity) bool) bool {
	cb2 := func(level int, id Identity) bool {
		if id == Anonymous {
			return false
		}
		if id == nil {
			return false
		}
		return cb(level, id)
	}

	if id, ok := v.(Identity); ok {
		if cb2(level, id) {
			return true
		}
	}

	if ipd, ok := v.(IdentityProvider); ok {
		if cb2(level, ipd.GetIdentity()) {
			return true
		}
	}

	if ipdgp, ok := v.(IdentityGroupProvider); ok {
		if cb2(level, ipdgp.GetIdentityGroup()) {
			return true
		}
	}

	return false
}

// IdentityProvider can be implemented by types that isn't itself and Identity,
// usually because they're not comparable/hashable.
type IdentityProvider interface {
	GetIdentity() Identity
}

// IdentityGroupProvider can be implemented by tightly connected types.
// Current use case is Resource transformation via Hugo Pipes.
type IdentityGroupProvider interface {
	GetIdentityGroup() Identity
}

// IsProbablyDependentProvider is an optional interface for Identity.
type IsProbablyDependentProvider interface {
	IsProbablyDependent(other Identity) bool
}

// GetDependencyManager returns the DependencyManager from v or nil if none found.
func GetDependencyManager(v any) Manager {
	switch vv := v.(type) {
	case Manager:
		return vv
	case types.Unwrapper:
		return GetDependencyManager(vv.Unwrapv())
	case DependencyManagerProvider:
		return vv.GetDependencyManager()
	}
	return nil
}

type DependencyManagerProvider interface {
	GetDependencyManager() Manager
}

type FindFirstManagerIdentityProvider interface {
	Identity
	FindFirstManagerIdentity() ManagerIdentity
}

func Unwrap(id Identity) Identity {
	switch t := id.(type) {
	case IdentityProvider:
		return t.GetIdentity()
	default:
		return id
	}
}

// IsProbablyDependencyProvider is an optional interface for Identity.
type IsProbablyDependencyProvider interface {
	IsProbablyDependency(other Identity) bool
}
