package identity

import (
	"fmt"
	"sort"
	"strings"
)

// Identities stores identity providers.
type Identities map[Identity]bool

func (ids Identities) AsSlice() []Identity {
	s := make([]Identity, len(ids))
	i := 0
	for v := range ids {
		s[i] = v
		i++
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].IdentifierBase() < s[j].IdentifierBase()
	})

	return s
}

func (ids Identities) String() string {
	var sb strings.Builder
	i := 0
	for id := range ids {
		sb.WriteString(fmt.Sprintf("[%s]", id.IdentifierBase()))
		if i < len(ids)-1 {
			sb.WriteString(", ")
		}
		i++
	}
	return sb.String()
}
