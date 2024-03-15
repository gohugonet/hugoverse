package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/query"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"github.com/gohugonet/hugoverse/pkg/form"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *Server) uploadContentsHandler(res http.ResponseWriter, req *http.Request) {
	order := query.Order(req)
	count, err := query.Count(req)
	if err != nil {
		s.Log.Errorf("Error parsing count: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}
	offset, err := query.Offset(req)
	if err != nil {
		s.Log.Errorf("Error parsing offset: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	b := &bytes.Buffer{}
	var total int
	var posts [][]byte

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="col s8">
						<div class="row">
							<div class="card-title col s7">Uploaded Items</div>
							<div class="col s5 input-field inline">
								<select class="browser-default __ponzu sort-order">
									<option value="DESC">New to Old</option>
									<option value="ASC">Old to New</option>
								</select>
								<label class="active">Sort:</label>
							</div>	
							<script>
								$(function() {
									var sort = $('select.__ponzu.sort-order');

									sort.on('change', function() {
										var path = window.location.pathname;
										var s = sort.val();

										window.location.replace(path + '?order=' + s);
									});

									var order = getParam('order');
									if (order !== '') {
										sort.val(order);
									}
									
								});
							</script>
						</div>
					</div>
					<form class="col s4" action="/admin/uploads/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all Upload fields" class="search"/>
							<input type="hidden" name="type" value="__uploads" />
						</div>
                    </form>	
					</div>`

	t := "__uploads"
	status := ""
	total, posts = db.Query(t, opts)

	pt := s.adminApp.UploadCreator()()
	p, ok := pt.(editor.Editable)
	if !ok {
		s.Log.Errorf("Error getting post type: %v", pt)
		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			return
		}

		res.Write(errView)
		return
	}

	for i := range posts {
		err := json.Unmarshal(posts[i], p)
		if err != nil {
			s.Log.Printf("Error unmarshal json into %s: %v", t, err)

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err := b.Write([]byte(post))
			if err != nil {
				s.Log.Errorf("Error writing post: %v", err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := s.adminView.Error500()
				if err != nil {
					s.Log.Errorf("Error rendering admin view: %v", err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write(post)
		if err != nil {
			s.Log.Errorf("Error writing post: %v", err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				s.Log.Errorf("Error rendering admin view: %v", err)
			}

			res.Write(errView)
			return
		}
	}

	html += `<ul class="posts row">`

	_, err = b.Write([]byte(`</ul>`))
	if err != nil {
		s.Log.Errorf("Error writing post: %v", err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	statusDisabled := "disabled"
	prevStatus := ""
	nextStatus := ""
	// total may be less than 10 (default count), so reset count to match total
	if total < count {
		count = total
	}
	// nothing previous to current list
	if offset == 0 {
		prevStatus = statusDisabled
	}
	// nothing after current list
	if (offset+1)*count >= total {
		nextStatus = statusDisabled
	}

	// set up pagination values
	urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s"
	prevURL := fmt.Sprintf(urlFmt, count, offset-1, order)
	nextURL := fmt.Sprintf(urlFmt, count, offset+1, order)
	start := 1 + count*offset
	end := start + count - 1

	if total < end {
		end = total
	}

	pagination := fmt.Sprintf(`
	<ul class="pagination row">
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_left</i></a></li>
		<li class="col s8">%d to %d of %d</li>
		<li class="col s2 waves-effect %s"><a href="%s"><i class="material-icons">chevron_right</i></a></li>
	</ul>
	`, prevStatus, prevURL, start, end, total, nextStatus, nextURL)

	// show indicator that a collection of items will be listed implicitly, but
	// that none are created yet
	if total < 1 {
		pagination = `
		<ul class="pagination row">
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_left</i></a></li>
			<li class="col s8">0 to 0 of 0</li>
			<li class="col s2 waves-effect disabled"><a href="#"><i class="material-icons">chevron_right</i></a></li>
		</ul>
		`
	}

	_, err = b.Write([]byte(pagination + `</div></div>`))
	if err != nil {
		s.Log.Errorf("Error writing post: %v", err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	script := `
	<script>
		$(function() {
			var del = $('.quick-delete-post.__ponzu span');
			del.on('click', function(e) {
				if (confirm("[Ponzu] Please confirm:\n\nAre you sure you want to delete this post?\nThis cannot be undone.")) {
					$(e.target).parent().submit();
				}
			});
		});

		// disable link from being clicked if parent is 'disabled'
		$(function() {
			$('ul.pagination li.disabled a').on('click', function(e) {
				e.preventDefault();
			});
		});
	</script>
	`

	btn := `<div class="col s3"><a href="/admin/edit/upload" class="btn new-post waves-effect waves-light">New Upload</a></div></div>`
	html = html + b.String() + script + btn

	adminView, err := s.adminView.SubView([]byte(html))
	if err != nil {
		s.Log.Errorf("Error rendering admin view: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

func (s *Server) editUploadHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		q := req.URL.Query()
		i := q.Get("id")
		t := "__uploads"

		post := s.adminApp.UploadCreator()()

		if i != "" {
			data, err := s.adminApp.GetUpload(i)
			if err != nil {
				s.Log.Errorf("Error getting upload: %v", err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := s.adminView.Error500()
				if err != nil {
					return
				}

				res.Write(errView)
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
				s.Log.Errorf("Error unmarshal json into %s: %v", t, err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := s.adminView.Error500()
				if err != nil {
					return
				}

				res.Write(errView)
				return
			}
		} else {
			it, ok := post.(content.Identifiable)
			if !ok {
				s.Log.Printf("Content type %s doesn't implement item.Identifiable", t)
				return
			}

			it.SetItemID(-1)
		}

		m, err := admin.Manage(post.(editor.Editable), t)
		if err != nil {
			s.Log.Errorf("Error rendering admin view: %v", err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		adminView, err := s.adminView.SubView(m)
		if err != nil {
			s.Log.Errorf("Error rendering admin view: %v", err)

			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
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

		t := req.FormValue("type")
		pt := "__uploads"
		ts := req.FormValue("timestamp")
		up := req.FormValue("updated")

		post := s.adminApp.UploadCreator()()

		// create a timestamp if one was not set
		if ts == "" {
			ts = fmt.Sprintf("%d", int64(time.Nanosecond)*time.Now().UTC().UnixNano()/int64(time.Millisecond))
			req.PostForm.Set("timestamp", ts)
		}

		if up == "" {
			req.PostForm.Set("updated", ts)
		}

		hook, ok := post.(content.Hookable)
		if !ok {
			s.Log.Printf("Type %s does not implement item.Hookable or embed item.Item.", pt)

			res.WriteHeader(http.StatusBadRequest)
			errView, err := s.adminView.Error400()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.BeforeSave(res, req)
		if err != nil {
			s.Log.Errorf("Error running BeforeSave method in editHandler for %s: %v", t, err)
			return
		}

		// StoreFiles has the SetUpload call (which is equivalent of SetContent in other handlers)
		urlPaths, err := s.StoreFiles(req)
		if err != nil {
			s.Log.Errorf("Error storing files: %v", err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		for name, urlPath := range urlPaths {
			req.PostForm.Set(name, urlPath)
		}

		req.PostForm, err = form.Convert(req.PostForm)
		if err != nil {
			s.Log.Errorf("Error converting post form: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				return
			}

			res.Write(errView)
			return
		}

		err = hook.AfterSave(res, req)
		if err != nil {
			s.Log.Printf("Error running AfterSave method in editHandler for %s: %v", t, err)
			return
		}

		scheme := req.URL.Scheme
		host := req.URL.Host
		redir := scheme + host + "/admin/uploads"
		http.Redirect(res, req, redir, http.StatusFound)

	case http.MethodPut:
		urlPaths, err := s.StoreFiles(req)
		if err != nil {
			s.Log.Errorf("Error storing files: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"data": [{"url": "` + urlPaths["file"] + `"}]}`))

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (s *Server) deleteUploadHandler(res http.ResponseWriter, req *http.Request) {
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
	t := "__uploads"

	if id == "" || t == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post := s.adminApp.UploadCreator()()
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

	err = hook.BeforeDelete(res, req)
	if err != nil {
		s.Log.Errorf("Error running BeforeDelete method in deleteHandler for %s: %v", t, err)
		return
	}

	// delete from file system, if good, we continue to delete
	// from database, if bad error 500
	err = s.deleteUploadFromDisk(id)
	if err != nil {
		s.Log.Errorf("Error deleting upload from disk: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.adminApp.DeleteUpload(id)
	if err != nil {
		s.Log.Errorf("Error deleting upload from database: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = hook.AfterDelete(res, req)
	if err != nil {
		s.Log.Errorf("Error running AfterDelete method in deleteHandler for %s: %v", t, err)
		return
	}

	redir := "/admin/uploads"
	http.Redirect(res, req, redir, http.StatusFound)
}

func (s *Server) uploadSearchHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	search := q.Get("q")
	status := q.Get("status")

	if search == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	t := "__uploads"

	posts, err := s.adminApp.AllUploads()
	if err != nil {
		s.Log.Errorf("Error getting all uploads: %v", err)
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}
	if posts == nil {
		s.Log.Printf("No uploads found.")
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	b := &bytes.Buffer{}

	pt := s.adminApp.UploadCreator()()
	p := pt.(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">Uploads Results</div>	
					<form class="col s4" action="/admin/uploads/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all Upload fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(search)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}

		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			log.Println("Error unmarshal search result json into", t, err, posts[i])

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err = b.Write([]byte(post))
			if err != nil {
				log.Println(err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := s.adminView.Error500()
				if err != nil {
					log.Println(err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write([]byte(post))
		if err != nil {
			log.Println(err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				log.Println(err)
			}

			res.Write(errView)
			return
		}
	}

	_, err = b.WriteString(`</ul></div></div>`)
	if err != nil {
		log.Println(err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			log.Println(err)
		}

		res.Write(errView)
		return
	}

	btn := `<div class="col s3"><a href="/admin/edit/upload" class="btn new-post waves-effect waves-light">New Upload</a></div></div>`
	html = html + b.String() + btn

	adminView, err := s.adminView.SubView([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}
