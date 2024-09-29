package handler

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"net/http"
)

func hide(res http.ResponseWriter, req *http.Request, it interface{}) bool {
	// check if should be hidden
	if h, ok := it.(content.Hideable); ok {
		err := h.Hide(res, req)
		if errors.Is(err, content.ErrAllowHiddenItem) {
			return false
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return true
		}

		res.WriteHeader(http.StatusNotFound)
		return true
	}

	return false
}
