package valueobject

import (
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"github.com/spf13/cast"
)

func Param(r contenthub.ResourceParamsProvider, fallback maps.Params, key any) (any, error) {
	keyStr, err := cast.ToStringE(key)
	if err != nil {
		return nil, err
	}

	if fallback == nil {
		return maps.GetNestedParam(keyStr, ".", r.Params())
	}

	return maps.GetNestedParam(keyStr, ".", r.Params(), fallback)
}
