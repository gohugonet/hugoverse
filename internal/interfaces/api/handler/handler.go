package handler

import (
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
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
	contentApp *entity.Content
	adminApp   *application.AdminServer
	adminView  *admin.View

	auth *auth.Auth
}

func New(log loggers.Logger, uploadDir string, db *database.Database,
	contentApp *entity.Content, adminApp *application.AdminServer) *Handler {

	adminView := admin.NewView(adminApp.Name(), contentApp.AllContentTypes())

	return &Handler{
		res: NewResponse(adminView),
		log: log,

		uploadDir: uploadDir,

		db:         db,
		contentApp: contentApp,
		adminApp:   adminApp,
		adminView:  adminView,

		auth: &auth.Auth{},
	}
}
