package api

import (
	"net/http"
)

func (s *Server) responseErr400(res http.ResponseWriter) error {
	res.WriteHeader(http.StatusBadRequest)
	errView, err := s.adminView.Error400()
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
	errView, err := s.adminView.Error500()
	if err != nil {
		return err
	}

	_, err = res.Write(errView)
	if err != nil {
		return err
	}
	return nil
}
