package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/maps"
	"html/template"
)

// MenuEntry represents a menu item defined in either Page front matter
// or in the site config.
type MenuEntry struct {
	// The menu entry configuration.
	MenuConfig

	// The menu containing this menu entry.
	Menu string

	// The URL value from front matter / config.
	ConfiguredURL string

	// Child entries.
	Children Menu
}

// Menu is a collection of menu entries.
type Menu []*MenuEntry

// Menus is a dictionary of menus.
type Menus map[string]Menu

// MenuConfig holds the configuration for a menu.
type MenuConfig struct {
	Identifier string
	Parent     string
	Name       string
	Pre        template.HTML
	Post       template.HTML
	URL        string
	PageRef    string
	Weight     int
	Title      string
	// User defined params.
	Params maps.Params
}
