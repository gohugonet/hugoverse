package api

import (
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"net/http"
)

func (s *Server) responseErr400(res http.ResponseWriter) error {
	res.WriteHeader(http.StatusBadRequest)
	errView, err := admin.Error400(s.adminApp.Name(), s.contentApp.AllContentTypes())
	if err != nil {
		return err
	}

	_, err = res.Write(errView)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) responseErr500(res http.ResponseWriter) error {
	res.WriteHeader(http.StatusInternalServerError)
	errView, err := admin.Error500(s.adminApp.Name(), s.contentApp.AllContentTypes())
	if err != nil {
		return err
	}

	_, err = res.Write(errView)
	if err != nil {
		return err
	}
	return nil
}
