package output

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/media"
)

// TODO, remove it

// DecodeFormats takes a list of output format configurations and merges those,
// in the order given, with the Hugo defaults as the last resort.
func DecodeFormats(mediaTypes media.Types) Formats {
	// Format could be modified by mediaTypes configuration
	// just make it simple for example
	fmt.Println(mediaTypes)

	f := make(Formats, len(DefaultFormats))
	copy(f, DefaultFormats)

	return f
}
