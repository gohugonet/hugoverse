package valueobject

import (
	"os"
	"strings"
)

var fileSeparator = string(os.PathSeparator)

type Mount struct {
	// Relative path in source repo, e.g. "scss".
	Source string

	// Relative target path, e.g. "assets/bootstrap/scss".
	Target string

	// Any file in this mount will be associated with this language.
	Lang string

	// Include only files matching the given Glob patterns (string or slice).
	IncludeFiles any

	// Exclude all files matching the given Glob patterns (string or slice).
	ExcludeFiles any
}

// Used as key to remove duplicates.
func (m Mount) key() string {
	return strings.Join([]string{m.Lang, m.Source, m.Target}, "/")
}

func (m Mount) Component() string {
	return strings.Split(m.Target, fileSeparator)[0]
}

func (m Mount) ComponentAndName() (string, string) {
	c, n, _ := strings.Cut(m.Target, fileSeparator)
	return c, n
}
