package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/query"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/form"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"strings"
)

func (s *Handler) ApiContentsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := s.contentApp.AllContentTypes()[t]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if hide(res, req, it()) {
		return
	}

	count, err := query.Count(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	offset, err := query.Offset(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	order := query.Order(req)

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	_, bb := s.db.Query(t+"__sorted", opts)
	var result []json.RawMessage
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := s.res.FmtJSON(result...)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// assert hookable
	get := it()
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

	s.res.Json(res, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}
}

func (s *Handler) ContentHandler(res http.ResponseWriter, req *http.Request) {
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
		s.log.Errorf("Error getting content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if post == nil {
		res.WriteHeader(http.StatusNotFound)
		s.log.Printf("Content not found: %s %s %s", t, id, status)
		return
	}

	p := pt()
	err = json.Unmarshal(post, p)
	if err != nil {
		s.log.Errorf("Error unmarshalling content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hide(res, req, p) {
		return
	}

	push(res, req, p, post)

	j, err := s.res.FmtJSON(json.RawMessage(post))
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

	s.res.Json(res, j)

	// hook after response
	err = hook.AfterAPIResponse(res, req, j)
	if err != nil {
		log.Println("[Response] error calling AfterAPIResponse:", err)
		return
	}
}

func (s *Handler) CreateContentHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		s.log.Errorf("Error parsing multipart form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	t := req.URL.Query().Get("type")
	if t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := s.contentApp.AllContentTypes()[t]
	if !found {
		s.log.Printf("Attempt to submit unknown type: %s from %s", t, req.RemoteAddr)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	post := p()

	ext, ok := post.(content.Createable)
	if !ok {
		s.log.Printf("Attempt to create non-createable type: %s from %s", t, req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	ts := timestamp.Now()
	req.PostForm.Set("timestamp", ts)
	req.PostForm.Set("updated", ts)

	urlPaths, err := s.StoreFiles(req)
	if err != nil {
		s.log.Errorf("Error storing files: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for name, urlPath := range urlPaths {
		req.PostForm.Set(name, urlPath)
	}

	req.PostForm, err = form.Convert(req.PostForm)
	if err != nil {
		s.log.Errorf("Error converting form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	hook, ok := post.(content.Hookable)
	if !ok {
		s.log.Printf("Attempt to create non-hookable type: %s from %s", t, req.RemoteAddr)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Let's be nice and make a proper item for the Hookable methods
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	dec.SetAliasTag("json")
	err = dec.Decode(post, req.PostForm)
	if err != nil {
		s.log.Printf("Error decoding post form for edit %s handler: %v", t, err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = hook.BeforeAPICreate(res, req)
	if err != nil {
		s.log.Errorf("Error calling BeforeCreate: %v", err)
		return
	}

	err = ext.Create(res, req)
	if err != nil {
		s.log.Errorf("Error calling Accept: %v", err)
		return
	}

	err = hook.BeforeSave(res, req)
	if err != nil {
		s.log.Errorf("Error calling BeforeSave: %v", err)
		return
	}

	// set specifier for db bucket in case content is/isn't Trustable
	var spec string

	// check if the content is Trustable should be auto-approved, if so the
	// content is immediately added to the public content API. If not, then it
	// is added to a "pending" list, only visible to Admins in the CMS and only
	// if the type implements editor.Mergable
	trusted, ok := post.(content.Trustable)
	if ok {
		err := trusted.AutoApprove(res, req)
		if err != nil {
			s.log.Errorf("Error calling AutoApprove: %v", err)
			return
		}
	} else {
		spec = "__pending"
		req.PostForm.Set("status", string(content.Pending))
	}

	req.PostForm.Set("namespace", t)
	s.log.Printf("PostForm: %+v", req.PostForm)

	id, err := s.contentApp.NewContent(t, req.PostForm)
	if err != nil {
		s.log.Errorf("Error calling SetContent: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set the target in the context so user can get saved value from db in hook
	ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%s", t, id))
	req = req.WithContext(ctx)

	err = hook.AfterSave(res, req)
	if err != nil {
		s.log.Errorf("Error calling AfterSave: %v", err)
		return
	}

	err = hook.AfterAPICreate(res, req)
	if err != nil {
		s.log.Errorf("Error calling AfterAccept: %v", err)
		return
	}

	// create JSON response to send data back to client
	var data map[string]interface{}
	if spec != "" {
		spec = strings.TrimPrefix(spec, "__")
		data = map[string]interface{}{
			"status": spec,
			"type":   t,
		}
	} else {
		spec = "public"
		data = map[string]interface{}{
			"id":     id,
			"status": spec,
			"type":   t,
		}
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
