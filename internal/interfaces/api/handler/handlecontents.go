package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Handler) ContentsHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	t := q.Get("type")
	if t == "" {
		if err := s.res.err400(res); err != nil {
			s.log.Errorf("Error response err 400: %s", err)
		}
		return
	}

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	contentType, ok := s.contentApp.AllContentTypes()[t]
	if !ok {
		if err := s.res.err400(res); err != nil {
			s.log.Errorf("Error response err 400: %s", err)
		}
		return
	}

	pt := contentType()

	p, ok := pt.(editor.Editable)
	if !ok {
		if err := s.res.err500(res); err != nil {
			s.log.Errorf("Error response err 500: %s", err)
		}
		return
	}

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
		} else {
			if err := s.res.err500(res); err != nil {
				s.log.Errorf("Error response err 500: %s", err)
			}
			return
		}
	}

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
		} else {
			if err := s.res.err500(res); err != nil {
				s.log.Errorf("Error response err 500: %s", err)
			}
			return
		}
	}

	opts := db.QueryOptions{
		Count:  count,
		Offset: offset,
		Order:  order,
	}

	status := q.Get("status")
	var specifier string
	if status == string(content.Public) || status == "" {
		specifier = "__sorted"
	} else if status == "pending" {
		specifier = "__pending"
	}

	b := &bytes.Buffer{}
	var total int
	var posts [][]byte

	html := s.adminView.Contents(t, status)

	var hasExt bool
	_, ok = pt.(content.Createable)
	if ok {
		hasExt = true
	}

	if hasExt {
		if status == "" {
			q.Set("status", "public")
		}

		// always start from top of results when changing public/pending
		q.Del("count")
		q.Del("offset")

		q.Set("status", "public")
		publicURL := req.URL.Path + "?" + q.Encode()

		q.Set("status", "pending")
		pendingURL := req.URL.Path + "?" + q.Encode()

		switch status {
		case "public", "":
			// get __sorted posts of type t from the db
			total, posts = s.db.Query(t+specifier, opts)

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<span class="active">Public</span>
					&nbsp;&vert;&nbsp;
					<a href="` + pendingURL + `">Pending</a>
				</div>`

			for i := range posts {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					s.log.Printf("Error unmarshal json into %s: %s", t, err, string(posts[i]))

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					_, err := b.Write([]byte(post))
					if err != nil {
						s.log.Errorf("Error writing post: %s", err)

						if err := s.res.err500(res); err != nil {
							s.log.Errorf("Error response err 500: %s", err)
						}
						return
					}

					continue
				}

				post := adminPostListItem(p, t, status)
				_, err = b.Write(post)
				if err != nil {
					s.log.Errorf("Error writing post: %s", err)

					if err := s.res.err500(res); err != nil {
						s.log.Errorf("Error response err 500: %s", err)
					}
					return
				}
			}

		case "pending":
			// get __pending posts of type t from the db
			total, posts = s.db.Query(t+"__pending", opts)

			html += `<div class="row externalable">
					<span class="description">Status:</span> 
					<a href="` + publicURL + `">Public</a>
					&nbsp;&vert;&nbsp;
					<span class="active">Pending</span>					
				</div>`

			for i := len(posts) - 1; i >= 0; i-- {
				err := json.Unmarshal(posts[i], &p)
				if err != nil {
					log.Println("Error unmarshal json into", t, err, string(posts[i]))

					post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
					_, err := b.Write([]byte(post))
					if err != nil {
						s.log.Errorf("Error writing post: %s", err)

						res.WriteHeader(http.StatusInternalServerError)
						errView, err := s.adminView.Error500()
						if err != nil {
							s.log.Errorf("Error response err 500: %s", err)
						}

						res.Write(errView)
						return
					}
					continue
				}

				post := adminPostListItem(p, t, status)
				_, err = b.Write(post)
				if err != nil {
					log.Println(err)

					res.WriteHeader(http.StatusInternalServerError)
					errView, err := s.adminView.Error500()
					if err != nil {
						s.log.Errorf("Error response err 500: %s", err)
					}

					res.Write(errView)
					return
				}
			}
		}

	} else {
		total, posts = s.db.Query(t+specifier, opts)

		for i := range posts {
			err := json.Unmarshal(posts[i], &p)
			if err != nil {
				log.Println("Error unmarshal json into", t, err, string(posts[i]))

				post := `<li class="col s12">Error decoding data. Possible file corruption.</li>`
				_, err := b.Write([]byte(post))
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
			_, err = b.Write(post)
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
	}

	html += `<ul class="posts row">`

	_, err = b.Write([]byte(`</ul>`))
	if err != nil {
		s.log.Errorf("Error response err 500: %s", err)

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
	urlFmt := req.URL.Path + "?count=%d&offset=%d&&order=%s&status=%s&type=%s"
	prevURL := fmt.Sprintf(urlFmt, count, offset-1, order, status, t)
	nextURL := fmt.Sprintf(urlFmt, count, offset+1, order, status, t)
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
		log.Println(err)

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

	btn := `<div class="col s3">
		<a href="/admin/edit?type=` + t + `" class="btn new-post waves-effect waves-light">
			New ` + t + `
		</a>`

	if _, ok := pt.(content.CSVFormattable); ok {
		btn += `<br/>
				<a href="/admin/contents/export?type=` + t + `&format=csv" class="green darken-4 btn export-post waves-effect waves-light">
					<i class="material-icons left">system_update_alt</i>
					CSV
				</a>`
	}

	html += b.String() + script + btn + `</div></div>`

	adminView, err := s.adminView.SubView([]byte(html))
	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(adminView)
}

// adminPostListItem is a helper to create the li containing a post.
// p is the asserted post as an Editable, t is the Type of the post.
// specifier is passed to append a name to a namespace like __pending
func adminPostListItem(e editor.Editable, typeName, status string) []byte {
	s, ok := e.(content.Sortable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement item.Sortable")
		post := `<li class="col s12">Error retreiving data. Your data type doesn't implement necessary interfaces. (item.Sortable)</li>`
		return []byte(post)
	}

	i, ok := e.(content.Identifiable)
	if !ok {
		log.Println("Content type", typeName, "doesn't implement item.Identifiable")
		post := `<li class="col s12">Error retreiving data. Your data type doesn't implement necessary interfaces. (item.Identifiable)</li>`
		return []byte(post)
	}

	// use sort to get other info to display in admin UI post list
	tsTime := time.Unix(int64(s.Time()/1000), 0)
	upTime := time.Unix(int64(s.Touch()/1000), 0)
	updatedTime := upTime.Format("01/02/06 03:04 PM")
	publishTime := tsTime.Format("01/02/06")

	cid := fmt.Sprintf("%d", i.ItemID())

	switch status {
	case "public", "":
		status = ""
	default:
		status = "__" + status
	}
	action := "/admin/edit/delete"
	link := `<a href="/admin/edit?type=` + typeName + `&status=` + strings.TrimPrefix(status, "__") + `&id=` + cid + `">` + i.String() + `</a>`
	if strings.HasPrefix(typeName, "__") {
		link = `<a href="/admin/edit/upload?id=` + cid + `">` + i.String() + `</a>`
		action = "/admin/edit/upload/delete"
	}

	post := `
			<li class="col s12">
				` + link + `
				<span class="post-detail">Updated: ` + updatedTime + `</span>
				<span class="publish-date right">` + publishTime + `</span>

				<form enctype="multipart/form-data" class="quick-delete-post __ponzu right" action="` + action + `" method="post">
					<span>Delete</span>
					<input type="hidden" name="id" value="` + cid + `" />
					<input type="hidden" name="type" value="` + typeName + status + `" />
				</form>
			</li>`

	return []byte(post)
}
