package entity

import "github.com/gohugonet/hugoverse/internal/domain/config/valueobject"

type Menu struct {
	Menus map[string][]valueobject.MenuConfig
}

func (m Menu) AllMenus() map[string][]valueobject.MenuConfig {
	return m.Menus
}
