package handler

import (
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"log"
	"net/http"
)

func (s *Handler) BuildContentHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	status := q.Get("status")

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, ok := s.contentApp.GetContentCreator(t)
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	p := pt()
	_, ok = p.(content.Buildable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Buildable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	t, err := s.contentApp.BuildTarget(t, id, status)
	if err != nil {
		s.log.Errorf("Error building: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = application.GenerateStaticSiteWithTarget(t)
	if err != nil {
		s.log.Errorf("Error building: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.res.Json(res, []byte("ok"))
}
