package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
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
			data, err := s.contentApp.GetContent(t, i, status)
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
			s.Log.Errorf("Error rendering admin view: %v", err)
			if err := s.responseErr500(res); err != nil {
				s.Log.Errorf("Error response err 500: %s", err)
			}
			return
		}

		adminView, err := s.adminView.SubView(m)
		if err != nil {
			s.Log.Errorf("Error rendering admin view: %v", err)
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
		pt := req.FormValue("type")
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

		ext, ok := post.(content.Createable)
		if !ok {
			s.Log.Errorf("[Create] type does not implement Createable:", pt, "from:", req.RemoteAddr)
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		err = ext.Create(res, req)
		if err != nil {
			s.Log.Errorf("[Create] error calling Create:", err)
			return
		}

		if cid == "-1" {
			err = hook.BeforeAdminCreate(res, req)
			if err != nil {
				s.Log.Errorf("Error running BeforeAdminCreate method in editHandler for:", pt, err)
				return
			}
		} else {
			err = hook.BeforeAdminUpdate(res, req)
			if err != nil {
				s.Log.Errorf("Error running BeforeAdminUpdate method in editHandler for:", pt, err)
				return
			}
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			s.Log.Errorf("Error running BeforeSave method in editHandler for:", pt, err)
			return
		}

		req.PostForm.Set("Name", pt)
		s.Log.Printf("PostForm: %+v", req.PostForm)

		if cid == "-1" {
			id, err := s.contentApp.NewContent(pt, req.PostForm)
			if err != nil {
				s.Log.Errorf("Error creating new content: %s", err)
				if err := s.responseErr500(res); err != nil {
					s.Log.Errorf("Error response err 500: %s", err)
				}
				return
			}

			cid = id
		} else {
			if err = s.contentApp.UpdateContent(pt, req.PostForm); err != nil {
				s.Log.Errorf("Error updating content: %s", err)
				if err := s.responseErr500(res); err != nil {
					s.Log.Errorf("Error response err 500: %s", err)
				}
				return
			}
		}

		if err := s.adminApp.InvalidateCache(); err != nil {
			s.Log.Errorf("Error invalidating cache: %s", err)
		}

		// set the target in the context so user can get saved value from db in hook
		ctx := context.WithValue(req.Context(), "target", fmt.Sprintf("%s:%s", pt, cid))
		req = req.WithContext(ctx)

		err = hook.AfterSave(res, req)
		if err != nil {
			s.Log.Errorf("Error running AfterSave method in editHandler for:", pt, err)
			return
		}

		if cid == "-1" {
			err = hook.AfterAdminCreate(res, req)
			if err != nil {
				s.Log.Errorf("Error running AfterAdminCreate method in editHandler for:", pt, err)
				return
			}
		} else {
			err = hook.AfterAdminUpdate(res, req)
			if err != nil {
				s.Log.Errorf("Error running AfterAdminUpdate method in editHandler for:", pt, err)
				return
			}
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		path := req.URL.Path
		redir := scheme + host + path + "?type=" + pt + "&id=" + cid

		if req.URL.Query().Get("status") == string(content.Pending) {
			redir += "&status=pending"
		}

		http.Redirect(res, req, redir, http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) deleteHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		s.Log.Errorf("Error parsing multipart form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	id := req.FormValue("id")
	t := req.FormValue("type")
	status := req.FormValue("status")
	ct := t

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	p, ok := s.contentApp.AllContentTypes()[ct]
	if !ok {
		s.Log.Printf("Type %s does not implement item.Hookable or embed item.Item.", t)
		res.WriteHeader(http.StatusBadRequest)
		errView, err := s.adminView.Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	post := p()
	hook, ok := post.(content.Hookable)
	if !ok {
		s.Log.Printf("Type %s does not implement item.Hookable or embed item.Item.", t)
		res.WriteHeader(http.StatusBadRequest)
		errView, err := s.adminView.Error400()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	data, err := s.contentApp.GetContent(t, id, status)
	if err != nil {
		s.Log.Printf("Error in db.Content %s:%s: %s", t, id, err)
		return
	}

	err = json.Unmarshal(data, post)
	if err != nil {
		log.Println("Error unmarshalling ", t, "=", id, err, " Hooks will be called on a zero-value.")
	}

	reject := req.URL.Query().Get("reject")
	if reject == "true" {
		err = hook.BeforeReject(res, req)
		if err != nil {
			log.Println("Error running BeforeReject method in deleteHandler for:", t, err)
			return
		}
	}

	err = hook.BeforeAdminDelete(res, req)
	if err != nil {
		log.Println("Error running BeforeAdminDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.BeforeDelete(res, req)
	if err != nil {
		log.Println("Error running BeforeDelete method in deleteHandler for:", t, err)
		return
	}

	err = s.contentApp.DeleteContent(t, id, status)
	if err != nil {
		s.Log.Errorf("Error in db.Content %s:%s: %s", t, id, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.adminApp.InvalidateCache(); err != nil {
		s.Log.Errorf("Error invalidating cache: %s", err)
	}

	err = hook.AfterDelete(res, req)
	if err != nil {
		s.Log.Errorf("Error running AfterDelete method in deleteHandler for:", t, err)
		return
	}

	err = hook.AfterAdminDelete(res, req)
	if err != nil {
		s.Log.Errorf("Error running AfterAdminDelete method in deleteHandler for:", t, err)
		return
	}

	if reject == "true" {
		err = hook.AfterReject(res, req)
		if err != nil {
			s.Log.Errorf("Error running AfterReject method in deleteHandler for:", t, err)
			return
		}
	}

	redir := strings.TrimSuffix(req.URL.Scheme+req.URL.Host+req.URL.Path, "/edit/delete")
	redir = redir + "/contents?type=" + ct
	http.Redirect(res, req, redir, http.StatusFound)
}
