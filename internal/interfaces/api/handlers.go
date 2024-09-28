package api

import (
	"github.com/gohugonet/hugoverse/internal/application"
	"net/http"
)

func (s *Server) registerContentHandler() {
	s.mux.HandleFunc("/api/contents", s.record.Collect(
		s.cors.Handle(s.comp.Gzip(s.handler.ApiContentsHandler))))
	s.mux.HandleFunc("/api/content", s.record.Collect(
		s.cors.Handle(s.comp.Gzip(s.handler.ContentHandler))))
	s.mux.HandleFunc("/api/content/create", s.record.Collect(
		s.cors.Handle(s.handler.CreateContentHandler)))

	s.mux.HandleFunc("/api/search", s.record.Collect(
		s.cors.Handle(s.comp.Gzip(s.handler.SearchContentHandler))))
}

func (s *Server) registerAdminHandler() {
	s.mux.HandleFunc("/admin", s.auth.Check(s.handler.AdminHandler))

	s.mux.HandleFunc("/admin/login", s.handler.LoginHandler)
	s.mux.HandleFunc("/admin/logout", s.handler.LogoutHandler)

	s.mux.HandleFunc("/admin/configure", s.auth.Check(s.handler.ConfigHandler))

	s.mux.HandleFunc("/admin/contents", s.auth.Check(s.handler.ContentsHandler))
	s.mux.HandleFunc("/admin/contents/search", s.auth.Check(s.handler.SearchHandler))

	s.mux.HandleFunc("/admin/edit", s.auth.Check(s.handler.EditHandler))
	s.mux.HandleFunc("/admin/edit/delete", s.auth.Check(s.handler.DeleteHandler))
	s.mux.HandleFunc("/admin/edit/approve", s.auth.Check(s.handler.ApproveContentHandler))

	s.mux.HandleFunc("/admin/uploads", s.auth.Check(s.handler.UploadContentsHandler))
	s.mux.HandleFunc("/admin/uploads/search", s.auth.Check(s.handler.UploadSearchHandler))
	s.mux.HandleFunc("/admin/edit/upload", s.auth.Check(s.handler.EditUploadHandler))
	s.mux.HandleFunc("/admin/edit/upload/delete", s.auth.Check(s.handler.DeleteUploadHandler))

	s.mux.HandleFunc("/admin/init", s.handler.InitHandler)

	staticDir := adminStaticDir()
	s.mux.Handle("/admin/static/", s.cache.Control(
		http.StripPrefix("/admin/static",
			http.FileServer(restrict(http.Dir(staticDir))))))

	uploadsDir := application.UploadDir()
	s.mux.Handle("/api/uploads/", s.record.Collect(s.cors.Handle(s.cache.Control(
		http.StripPrefix("/api/uploads/",
			http.FileServer(restrict(http.Dir(uploadsDir))))))))

}
