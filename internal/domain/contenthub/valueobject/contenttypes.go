package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/media"
	"path/filepath"
	"strings"
)

var DefaultContentTypes ContentTypes

func SetupDefaultContentTypes() {
	DefaultContentTypes = ContentTypes{
		HTML:             media.Builtin.HTMLType,
		Markdown:         media.Builtin.MarkdownType,
		AsciiDoc:         media.Builtin.AsciiDocType,
		Pandoc:           media.Builtin.PandocType,
		ReStructuredText: media.Builtin.ReStructuredTextType,
		EmacsOrgMode:     media.Builtin.EmacsOrgModeType,
	}

	DefaultContentTypes.setup()
}

// ContentTypes holds the media types that are considered content in Hugo.
type ContentTypes struct {
	HTML             media.Type
	Markdown         media.Type
	AsciiDoc         media.Type
	Pandoc           media.Type
	ReStructuredText media.Type
	EmacsOrgMode     media.Type

	// Created in init().
	types        media.Types
	extensionSet map[string]bool
}

func (t *ContentTypes) setup() {
	t.types = media.Types{t.HTML, t.Markdown, t.AsciiDoc, t.Pandoc, t.ReStructuredText, t.EmacsOrgMode}
	t.extensionSet = make(map[string]bool)
	for _, mt := range t.types {
		for _, suffix := range mt.Suffixes() {
			t.extensionSet[suffix] = true
		}
	}
}

func (t ContentTypes) IsContentSuffix(suffix string) bool {
	return t.extensionSet[suffix]
}

// IsContentFile returns whether the given filename is a content file.
func (t ContentTypes) IsContentFile(filename string) bool {
	return t.IsContentSuffix(strings.TrimPrefix(filepath.Ext(filename), "."))
}

// IsIndexContentFile returns whether the given filename is an index content file.
func (t ContentTypes) IsIndexContentFile(filename string) bool {
	if !t.IsContentFile(filename) {
		return false
	}

	base := filepath.Base(filename)

	return strings.HasPrefix(base, "index.") || strings.HasPrefix(base, "_index.")
}

// IsHTMLSuffix returns whether the given suffix is a HTML media type.
func (t ContentTypes) IsHTMLSuffix(suffix string) bool {
	for _, s := range t.HTML.Suffixes() {
		if s == suffix {
			return true
		}
	}
	return false
}

// Types is a slice of media types.
func (t ContentTypes) Types() media.Types {
	return t.types
}

// FromTypes creates a new ContentTypes updated with the values from the given Types.
func (t ContentTypes) FromTypes(types media.Types) ContentTypes {
	if tt, ok := types.GetByType(t.HTML.Type); ok {
		t.HTML = tt
	}
	if tt, ok := types.GetByType(t.Markdown.Type); ok {
		t.Markdown = tt
	}
	if tt, ok := types.GetByType(t.AsciiDoc.Type); ok {
		t.AsciiDoc = tt
	}
	if tt, ok := types.GetByType(t.Pandoc.Type); ok {
		t.Pandoc = tt
	}
	if tt, ok := types.GetByType(t.ReStructuredText.Type); ok {
		t.ReStructuredText = tt
	}
	if tt, ok := types.GetByType(t.EmacsOrgMode.Type); ok {
		t.EmacsOrgMode = tt
	}

	t.setup()

	return t
}
