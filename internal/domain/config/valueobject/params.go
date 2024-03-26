package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/config"

func SetParams(p, pp config.Params) {
	for k, v := range pp {
		vv, found := p[k]
		if !found {
			p[k] = v
		} else {
			switch vvv := vv.(type) {
			case config.Params:
				if pv, ok := v.(config.Params); ok {
					SetParams(vvv, pv)
				} else {
					p[k] = v
				}
			default:
				p[k] = v
			}
		}
	}
}
