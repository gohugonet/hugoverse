package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/compare"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"html/template"
	"sort"
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

	// TODO, removed
	Page *MenuEntry
}

// HasChildren returns whether this menu item has any children.
func (m *MenuEntry) HasChildren() bool {
	return m.Children != nil
}

// KeyName returns the key used to identify this menu entry.
func (m *MenuEntry) KeyName() string {
	if m.Identifier != "" {
		return m.Identifier
	}
	return m.Name
}

// Menu is a collection of menu entries.
type Menu []*MenuEntry

func (m Menu) Add(me *MenuEntry) Menu {
	m = append(m, me)
	m.Sort()
	return m
}

func (m Menu) Sort() Menu {
	menuEntryBy(defaultMenuEntrySort).Sort(m)
	return m
}

func (m Menu) Name() string {
	if len(m) > 0 {
		return m[0].Name
	}
	return ""
}

func (m Menu) URL() string {
	if len(m) > 0 {
		return m[0].URL
	}
	return ""
}

type menuEntryBy func(m1, m2 *MenuEntry) bool

func (by menuEntryBy) Sort(menu Menu) {
	ms := &menuSorter{
		menu: menu,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Stable(ms)
}

type menuSorter struct {
	menu Menu
	by   menuEntryBy
}

func (ms *menuSorter) Len() int      { return len(ms.menu) }
func (ms *menuSorter) Swap(i, j int) { ms.menu[i], ms.menu[j] = ms.menu[j], ms.menu[i] }

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (ms *menuSorter) Less(i, j int) bool { return ms.by(ms.menu[i], ms.menu[j]) }

var defaultMenuEntrySort = func(m1, m2 *MenuEntry) bool {
	if m1.Weight == m2.Weight {
		c := compare.Strings(m1.Name, m2.Name)
		if c == 0 {
			return m1.Identifier < m2.Identifier
		}
		return c < 0
	}

	if m2.Weight == 0 {
		return true
	}

	if m1.Weight == 0 {
		return false
	}

	return m1.Weight < m2.Weight
}

const MenusAfter = "after"
const MenusBefore = "before"

// Menus is a dictionary of menus.
type Menus map[string]Menu

func NewEmptyMenus() Menus {
	return Menus{
		MenusBefore: Menu{},
		MenusAfter:  Menu{},
	}
}

func (m Menus) HasSubMenu(me *MenuEntry) bool {
	if sm, ok := m[me.Name]; ok {
		if len(sm) == 1 && sm[0].Name == me.Name {
			return false
		}
		return len(sm) > 1
	}
	return false
}

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
