package identity

import (
	"fmt"
	"sync"
)

type ManagerOption func(m *identityManager)

// NewIdentityManager creates a new Manager.
func NewManager(name string, opts ...ManagerOption) Manager {
	idm := &identityManager{
		Identity: Anonymous,
		name:     name,
		ids:      Identities{},
	}

	for _, o := range opts {
		o(idm)
	}

	return idm
}

type identityManager struct {
	Identity

	// Only used for debugging.
	name string

	// mu protects _changes_ to this manager,
	// reads currently assumes no concurrent writes.
	mu         sync.RWMutex
	ids        Identities
	forEachIds []ForEeachIdentityProvider

	// Hooks used in debugging.
	onAddIdentity func(id Identity)
}

func (im *identityManager) AddIdentity(ids ...Identity) {
	im.mu.Lock()

	for _, id := range ids {
		if id == nil || id == Anonymous {
			continue
		}
		if _, found := im.ids[id]; !found {
			if im.onAddIdentity != nil {
				im.onAddIdentity(id)
			}
			im.ids[id] = true
		}
	}
	im.mu.Unlock()
}

func (im *identityManager) AddIdentityForEach(ids ...ForEeachIdentityProvider) {
	im.mu.Lock()
	im.forEachIds = append(im.forEachIds, ids...)
	im.mu.Unlock()
}

func (im *identityManager) ContainsIdentity(id Identity) FinderResult {
	if im.Identity != Anonymous && id == im.Identity {
		return FinderFound
	}

	f := NewFinder(FinderConfig{Exact: true})
	r := f.Contains(id, im, -1)

	return r
}

// Managers are always anonymous.
func (im *identityManager) GetIdentity() Identity {
	return im.Identity
}

func (im *identityManager) Reset() {
	im.mu.Lock()
	im.ids = Identities{}
	im.mu.Unlock()
}

func (im *identityManager) GetDependencyManagerForScope(int) Manager {
	return im
}

func (im *identityManager) String() string {
	return fmt.Sprintf("IdentityManager(%s)", im.name)
}

func (im *identityManager) forEeachIdentity(fn func(id Identity) bool) bool {
	// The absence of a lock here is deliberate. This is currently only used on server reloads
	// in a single-threaded context.
	for id := range im.ids {
		if fn(id) {
			return true
		}
	}
	for _, fe := range im.forEachIds {
		if fe.ForEeachIdentity(fn) {
			return true
		}
	}
	return false
}
