package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/host/entity"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func NewHost(log loggers.Logger) (*entity.Host, error) {
	netlify, err := entity.NewNetlify(log)
	if err != nil {
		return nil, err
	}

	return &entity.Host{
		Netlify: netlify,
	}, nil
}
