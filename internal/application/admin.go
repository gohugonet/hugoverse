package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/factory"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"github.com/nilslice/jwt"
)

type AdminServer struct {
	admin.Admin
}

func NewAdminServer(db repository.Repository) (*AdminServer, error) {
	a, err := factory.NewAdmin(db)
	if err != nil {
		return nil, err
	}

	if a.ClientSecret() != "" {
		jwt.Secret([]byte(a.ClientSecret()))
	}
	err = a.InvalidateCache()
	if err != nil {
		return nil, err
	}

	return &AdminServer{
		Admin: a,
	}, nil
}
