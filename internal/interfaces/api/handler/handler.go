package handler

import (
	"github.com/mdfriday/hugoverse/internal/application"
	adminEntity "github.com/mdfriday/hugoverse/internal/domain/admin/entity"
	contentEntity "github.com/mdfriday/hugoverse/internal/domain/content/entity"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/admin"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/auth"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/database"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"html/template"
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

	adminView := &admin.View{
		Logo:       adminApp.Name(),
		Types:      contentApp.AllContentTypes(),
		AdminTypes: contentApp.AllAdminTypes(),
		AdminEmail: adminApp.Conf.AdminEmail,
		Subview:    template.HTML(""),
	}

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
