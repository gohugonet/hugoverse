package valueobject

import "github.com/disintegration/gift"

const filterAPIVersion = 0

type filter struct {
	Options filterOpts
	gift.Filter
}

// For cache-busting.
type filterOpts struct {
	Version int
	Vals    any
}

func newFilterOpts(vals ...any) filterOpts {
	return filterOpts{
		Version: filterAPIVersion,
		Vals:    vals,
	}
}
