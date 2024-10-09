package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (s *Handler) SearchContentHandler(res http.ResponseWriter, req *http.Request) {
	qs := req.URL.Query()
	t := qs.Get("type")
	if t == "" {
		s.log.Printf("Type must be set")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	it, ok := s.contentApp.AllContentTypes()[t]
	if !ok {
		s.log.Printf("Type %s not found", t)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if hide(res, req, it()) {
		return
	}

	q, err := url.QueryUnescape(qs.Get("q"))
	if err != nil {
		s.log.Errorf("Error unescaping query: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// q must be set
	if q == "" {
		s.log.Errorf("Query must be set")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	count, err := strconv.Atoi(qs.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if qs.Get("count") == "" {
			count = 10
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	offset, err := strconv.Atoi(qs.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if qs.Get("offset") == "" {
			offset = 0
		} else {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// execute search for query provided, if no index for type send 404
	indices, err := s.contentApp.Search.TypeQuery(t, q, count, offset)
	if errors.Is(err, content.ErrNoIndex) {
		s.log.Errorf("Index for type %s not found", t)
		res.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		s.log.Errorf("Error searching for type %s: %v", t, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// respond with json formatted results
	bb, err := s.contentApp.GetContents(indices)
	if err != nil {
		s.log.Errorf("Error getting content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// if we have matches, push the first as its matched by relevance
	if len(bb) > 0 {
		push(res, req, it(), bb[0])
	}

	var result []json.RawMessage
	for i := range bb {
		result = append(result, bb[i])
	}

	j, err := s.res.FmtJSON(result...)
	if err != nil {
		s.log.Errorf("Error formatting json: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err = omit(res, req, it(), j)
	if err != nil {
		s.log.Errorf("Error formatting json: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.res.Json(res, j)
}

func (s *Handler) SearchHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	query := q.Get("q")
	status := q.Get("status")
	var specifier string

	if t == "" || query == "" {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	if status == "pending" {
		specifier = "__" + status
	}

	posts := s.db.AllContent(t + specifier)
	b := &bytes.Buffer{}
	pt, ok := s.contentApp.AllContentTypes()[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	post := pt()

	p := post.(editor.Editable)

	html := `<div class="col s9 card">		
					<div class="card-content">
					<div class="row">
					<div class="card-title col s7">` + t + ` Results</div>	
					<form class="col s4" action="/admin/contents/search" method="get">
						<div class="input-field post-search inline">
							<label class="active">Search:</label>
							<i class="right material-icons search-icon">search</i>
							<input class="search" name="q" type="text" placeholder="Within all ` + t + ` fields" class="search"/>
							<input type="hidden" name="type" value="` + t + `" />
							<input type="hidden" name="status" value="` + status + `" />
						</div>
                    </form>	
					</div>
					<ul class="posts row">`

	for i := range posts {
		// skip posts that don't have any matching search criteria
		match := strings.ToLower(query)
		all := strings.ToLower(string(posts[i]))
		if !strings.Contains(all, match) {
			continue
		}

		err := json.Unmarshal(posts[i], &p)
		if err != nil {
			s.log.Printf("Error unmarshal search result json into %s with err: %v, content: %v", t, err, posts[i])

			post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
			_, err = b.Write([]byte(post))
			if err != nil {
				s.log.Errorf("[admin] Error: %v", err)

				res.WriteHeader(http.StatusInternalServerError)
				errView, err := s.adminView.Error500()
				if err != nil {
					s.log.Errorf("[admin] Error: %v", err)
				}

				res.Write(errView)
				return
			}
			continue
		}

		post := adminPostListItem(p, t, status)
		_, err = b.Write([]byte(post))
		if err != nil {
			s.log.Errorf("[admin] Error: %v", err)

			res.WriteHeader(http.StatusInternalServerError)
			errView, err := s.adminView.Error500()
			if err != nil {
				s.log.Errorf("[admin] Error: %v", err)
			}

			res.Write(errView)
			return
		}
	}

	_, err := b.WriteString(`</ul></div></div>`)
	if err != nil {
		s.log.Errorf("[admin] Error: %v", err)

		res.WriteHeader(http.StatusInternalServerError)
		errView, err := s.adminView.Error500()
		if err != nil {
			s.log.Errorf("[admin] Error: %v", err)
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

	btn := `<div class="col s3">
		<a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">
			New ` + t + `
		</a>`

	html += b.String() + script + btn + `</div></div>`

	adminView, err := s.adminView.SubView([]byte(html))
	if err != nil {
		s.log.Errorf("[admin] Error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}
