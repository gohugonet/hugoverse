package handler

import (
	"encoding/json"
	"errors"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"net/http"
)

func (s *Handler) handlerError(res http.ResponseWriter, req *http.Request, err error) {
	var fe herrors.FileError
	if errors.As(err, &fe) {
		res.WriteHeader(http.StatusBadRequest)

		jsonBytes, err := json.Marshal(fe.Error())
		if err != nil {
			s.log.Errorf("Error marshalling token when handling error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		j, err := s.res.FmtJSON(jsonBytes)
		if err != nil {
			s.log.Errorf("Error formatting JSON when handling error: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.res.Json(res, j)

		return
	}

	res.WriteHeader(http.StatusInternalServerError)
	_, err = res.Write([]byte(err.Error()))
	if err != nil {
		s.log.Errorf("Error writing response: %v", err)
	}

	return
}
