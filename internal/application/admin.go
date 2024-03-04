package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/factory"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
)

type AdminServer struct {
	admin.Admin
}

func NewAdminServer(db repository.Repository) *AdminServer {
	return &AdminServer{
		Admin: factory.NewAdmin(db),
	}
}
