package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"strings"
)

func (s *Server) editHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		i := q.Get("id")
		t := q.Get("type")
		status := q.Get("status")

		contentType, ok := s.contentApp.AllContentTypes()[t]
		if !ok {
			fmt.Fprintf(res, content.ErrTypeNotRegistered.Error(), t)
			return
		}
		post := contentType()

		if i != "" {
			if status == "pending" {
				t = t + "__pending"
			}

			data, err := db.Content(t + ":" + i)
			if err != nil {
				if err := s.responseErr500(res); err != nil {
					s.Log.Errorf("Error response err 500: %s", err)
				}
				return
			}

			if len(data) < 1 || data == nil {
				res.WriteHeader(http.StatusNotFound)
				errView, err := s.adminView.Error404()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}

			err = json.Unmarshal(data, post)
			if err != nil {
				if err := s.responseErr500(res); err != nil {
					s.Log.Errorf("Error response err 500: %s", err)
				}
				return
			}
		} else {
			item, ok := post.(content.Identifiable)
			if !ok {
				s.Log.Printf("Content type %s doesn't implement item.Identifiable", t)
				return
			}

			item.SetItemID(-1)
		}

		m, err := admin.Manage(post.(editor.Editable), t)
		if err != nil {
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		adminView, err := s.adminView.SubView(m)
		if err != nil {
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
		if err != nil {
			s.Log.Errorf("Error parsing multipart form: %v", err)
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		cid := req.FormValue("id")
		t := req.FormValue("type")
		ts := req.FormValue("timestamp")
		up := req.FormValue("updated")

		// create a timestamp if one was not set
		if ts == "" {
			ts = timestamp.Now()
			req.PostForm.Set("timestamp", ts)
		}

		if up == "" {
			req.PostForm.Set("updated", ts)
		}

		urlPaths, err := s.StoreFiles(req)
		if err != nil {
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		for name, urlPath := range urlPaths {
			req.PostForm.Set(name, urlPath)
		}

		// check for any multi-value fields (ex. checkbox fields)
		// and correctly format for db storage. Essentially, we need
		// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
		fieldOrderValue := make(map[string]map[string][]string)
		for k, v := range req.PostForm {
			if strings.Contains(k, ".") {
				fo := strings.Split(k, ".")

				// put the order and the field value into map
				field := string(fo[0])
				order := string(fo[1])
				if len(fieldOrderValue[field]) == 0 {
					fieldOrderValue[field] = make(map[string][]string)
				}

				// orderValue is 0:[?type=Thing&id=1]
				orderValue := fieldOrderValue[field]
				orderValue[order] = v
				fieldOrderValue[field] = orderValue

				// discard the post form value with name.N
				req.PostForm.Del(k)
			}

		}

		// add/set the key & value to the post form in order
		for f, ov := range fieldOrderValue {
			for i := 0; i < len(ov); i++ {
				position := fmt.Sprintf("%d", i)
				fieldValue := ov[position]

				if req.PostForm.Get(f) == "" {
					for i, fv := range fieldValue {
						if i == 0 {
							req.PostForm.Set(f, fv)
						} else {
							req.PostForm.Add(f, fv)
						}
					}
				} else {
					for _, fv := range fieldValue {
						req.PostForm.Add(f, fv)
					}
				}
			}
		}

		pt := t
		if strings.Contains(t, "__") {
			pt = strings.Split(t, "__")[0]
		}

		p, ok := s.contentApp.AllContentTypes()[pt]
		if !ok {
			if err := s.responseErr400(res); err != nil {
				s.Log.Errorf("Error response err 400: %s", err)
			}
			return
		}

		post := p()
		hook, ok := post.(content.Hookable)
		if !ok {
			if err := s.responseErr400(res); err != nil {
				s.Log.Errorf("Error response err 400: %s", err)
			}
			return
		}

		// Let's be nice and make a proper item for the Hookable methods
		dec := schema.NewDecoder()
		dec.IgnoreUnknownKeys(true)
		dec.SetAliasTag("json")
		err = dec.Decode(post, req.PostForm)
		if err != nil {
			if err := s.responseErr400(res); err != nil {
				s.Log.Errorf("Error response err 400: %s", err)
			}
			return
		}

		if cid == "-1" {
			err = hook.BeforeAdminCreate(res, req)
			if err != nil {
				log.Println("Error running BeforeAdminCreate method in editHandler for:", t, err)
				return
			}
		} else {
			err = hook.BeforeAdminUpdate(res, req)
			if err != nil {
				log.Println("Error running BeforeAdminUpdate method in editHandler for:", t, err)
				return
			}
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			log.Println("Error running BeforeSave method in editHandler for:", t, err)
			return
		}

		id, err := s.contentApp.NewContent(pt, req.PostForm)
		if err != nil {
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%s", t, id))
		req = req.WithContext(ctx)

		err = hook.AfterSave(res, req)
		if err != nil {
			log.Println("Error running AfterSave method in editHandler for:", t, err)
			return
		}

		if cid == "-1" {
			err = hook.AfterAdminCreate(res, req)
			if err != nil {
				log.Println("Error running AfterAdminUpdate method in editHandler for:", t, err)
				return
			}
		} else {
			err = hook.AfterAdminUpdate(res, req)
			if err != nil {
				log.Println("Error running AfterAdminUpdate method in editHandler for:", t, err)
				return
			}
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		path := req.URL.Path
		sid := fmt.Sprintf("%d", id)
		redir := scheme + host + path + "?type=" + pt + "&id=" + sid

		if req.URL.Query().Get("status") == "pending" {
			redir += "&status=pending"
		}

		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
