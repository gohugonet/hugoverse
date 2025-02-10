package handler

import (
	"encoding/json"
	"github.com/mdfriday/hugoverse/internal/domain/content"
	"net/http"
)

func (s *Handler) HashHandler(res http.ResponseWriter, req *http.Request) {
	s.getContentByHash(res, req)
}

func (s *Handler) getContentByHash(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	status := q.Get("status")
	hash := q.Get("hash")

	if t == "" || hash == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, ok := s.contentApp.GetContentCreator(t)
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}
	p := pt()

	_, ok = p.(content.Hashable)
	if !ok {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	post, err := s.contentApp.GetContentByHash(t, hash, status)
	if err != nil {
		s.log.Errorf("Error getting content by hash %s: %v", hash, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if post == nil {
		res.WriteHeader(http.StatusNotFound)
		s.log.Debugf("Content not found: %s %s %s", t, hash, status)
		return
	}

	err = json.Unmarshal(post, p)
	if err != nil {
		s.log.Errorf("Error unmarshalling content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"id":   p.(content.Identifier).ID(),
		"type": t,
	}

	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			data,
		},
	}

	j, err := json.Marshal(resp)
	if err != nil {
		s.log.Errorf("Error marshalling response to JSON: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	_, err = res.Write(j)
	if err != nil {
		s.log.Errorf("Error writing response: %v", err)
		return
	}
}
