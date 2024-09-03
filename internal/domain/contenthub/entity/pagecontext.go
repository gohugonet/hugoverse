package entity

import "github.com/gohugonet/hugoverse/pkg/text"

// pageContext provides contextual information about this page, for error
// logging and similar.
type pageContext interface {
	posOffset(offset int) text.Position
}
