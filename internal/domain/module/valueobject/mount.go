package valueobject

import (
	"github.com/mdfriday/hugoverse/pkg/paths"
	"strings"
)

var fileSeparator = string(paths.FilePathSeparator)

type Mount struct {
	// Relative path in source repo, e.g. "scss".
	SourcePath string

	// Relative target path, e.g. "assets/bootstrap/scss".
	TargetPath string

	// Any file in this mount will be associated with this language.
	Language string
}

func (m Mount) Marshal() string {
	return strings.Join([]string{m.Language, m.SourcePath, m.TargetPath}, "/")
}

func (m Mount) Component() string {
	return strings.Split(m.TargetPath, fileSeparator)[0]
}

func (m Mount) ComponentAndName() (string, string) {
	c, n, _ := strings.Cut(m.TargetPath, fileSeparator)
	return c, n
}

func (m Mount) Source() string { return m.SourcePath }
func (m Mount) Target() string { return m.TargetPath }
func (m Mount) Lang() string   { return m.Language }
