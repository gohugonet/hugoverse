package api

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"log"
	"net/http"
)

func (s *Server) registerContentHandler() {
	s.mux.HandleFunc("/api/contents", Record(s.CORS(s.Gzip(s.contentHandler))))

	s.mux.HandleFunc("/api/search", Record(s.CORS(s.Gzip(s.searchContentHandler))))
}

func (s *Server) contentHandler(res http.ResponseWriter, req *http.Request) {
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

	post, err := s.contentApp.GetContent(t, id, status)
	if err != nil {
		s.Log.Errorf("Error getting content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if post == nil {
		res.WriteHeader(http.StatusNotFound)
		s.Log.Printf("Content not found: %s %s %s", t, id, status)
		return
	}

	p := pt()
	err = json.Unmarshal(post, p)
	if err != nil {
		s.Log.Errorf("Error unmarshalling content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hide(res, req, p) {
		return
	}

	push(res, req, p, post)

	j, err := fmtJSON(json.RawMessage(post))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, p, j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// assert hookable
	get := p
	hook, ok := get.(content.Hookable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Hookable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// hook before response
	j, err = hook.BeforeAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling BeforeAPIResponse:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	sendData(res, req, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}
}
