package handler

import (
	"github.com/gohugonet/hugoverse/internal/application"
	adminEntity "github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	contentEntity "github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/auth"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/database"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

type Handler struct {
	res *Response
	log loggers.Logger

	uploadDir string

	db         *database.Database
	contentApp *contentEntity.Content
	adminApp   *adminEntity.Admin
	adminView  *admin.View

	auth *auth.Auth
}

func New(log loggers.Logger, db *database.Database,
	contentApp *contentEntity.Content, adminApp *adminEntity.Admin) *Handler {

	adminView := admin.NewView(adminApp.Name(), contentApp.AllContentTypes())

	return &Handler{
		res: NewResponse(adminView),
		log: log,

		uploadDir: application.UploadDir(),

		db:         db,
		contentApp: contentApp,
		adminApp:   adminApp,
		adminView:  adminView,

		auth: &auth.Auth{},
	}
}
