package entity

import "github.com/gohugonet/hugoverse/internal/domain/config/valueobject"

type Language struct {
	Configs     map[string]valueobject.LanguageConfig
	RootConfigs map[string]valueobject.RootConfig
}
