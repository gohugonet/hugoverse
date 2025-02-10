package valueobject

import (
	"github.com/mdfriday/hugoverse/pkg/identity"
	"sync"
)

type WhatChanged struct {
	mu sync.Mutex

	IdentitySet identity.Identities
}

func (w *WhatChanged) Add(ids ...identity.Identity) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, id := range ids {
		w.IdentitySet[id] = true
	}
}

func (w *WhatChanged) Changes() []identity.Identity {
	if w == nil || w.IdentitySet == nil {
		return nil
	}
	return w.IdentitySet.AsSlice()
}
