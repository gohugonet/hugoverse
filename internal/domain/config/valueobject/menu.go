package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"html/template"
)

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

func DecodeMenuConfig(cfg config.Provider) (map[string][]MenuConfig, error) {
	return decodeMenuConfig(cfg)
}

func decodeMenuConfig(cfg config.Provider) (map[string][]MenuConfig, error) {
	menuMap := make(map[string][]MenuConfig)

	if !cfg.IsSet("menu") {
		return menuMap, nil
	}

	m := cfg.Get("menu")

	menus, err := maps.ToStringMapE(m)
	if err != nil {
		return menuMap, err
	}

	menus = maps.CleanConfigStringMap(menus)

	for name, menu := range menus {
		m, err := cast.ToSliceE(menu)
		if err != nil {
			return menuMap, err
		} else {
			for _, entry := range m {
				var menuConfig MenuConfig
				if err := mapstructure.WeakDecode(entry, &menuConfig); err != nil {
					return menuMap, err
				}
				maps.PrepareParams(menuConfig.Params)

				if menuMap[name] == nil {
					menuMap[name] = []MenuConfig{}
				}

				menuMap[name] = append(menuMap[name], menuConfig)
			}
		}
	}

	return menuMap, nil
}
