package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"log"
	"net/http"
)

type Author struct {
	Item

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	SiteURL   string `json:"site_url"`
	Avatar    string `json:"avatar"`
}

// MarshalEditor writes a buffer of html to edit a Song within the CMS
// and implements editor.Editable
func (s *Author) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(s,
		editor.Field{
			View: editor.Input("FirstName", s, map[string]string{
				"label":       "First Name",
				"type":        "text",
				"placeholder": "Enter the first name here",
			}),
		},
		editor.Field{
			View: editor.Input("LastName", s, map[string]string{
				"label":       "Last Name",
				"type":        "text",
				"placeholder": "Enter the last name here",
			}),
		},
		editor.Field{
			View: editor.Input("Email", s, map[string]string{
				"label":       "Email",
				"type":        "text",
				"placeholder": "Enter the email here",
			}),
		},
		editor.Field{
			View: editor.Input("SiteURL", s, map[string]string{
				"label":       "SiteURL",
				"type":        "text",
				"placeholder": "Enter the site URL here",
			}),
		},
		editor.Field{
			View: editor.File("Avatar", s, map[string]string{
				"label":       "Avatar",
				"placeholder": "Upload the Avatar here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to render Author editor view: %s", err.Error())
	}

	return view, nil
}

// String defines the display name of a Song in the CMS list-view
func (s *Author) String() string { return s.FirstName + " " + s.LastName }

// Create implements api.Createable, and allows external POST requests from clients
// to add content as long as the request contains the json tag names of the Song
// struct fields, and is multipart encoded
func (s *Author) Create(res http.ResponseWriter, req *http.Request) error {
	// do form data validation for required fields
	required := []string{
		"first_name",
		"last_name",
	}

	for _, r := range required {
		if req.PostFormValue(r) == "" {
			err := fmt.Errorf("request missing required field: %s", r)
			return err
		}
	}

	return nil
}

// BeforeAPICreate is only called if the Song type implements api.Createable
// It is called before Create, and returning an error will cancel the request
// causing the system to reject the data sent in the POST
func (s *Author) BeforeAPICreate(res http.ResponseWriter, req *http.Request) error {
	// do initial user authentication here on the request, checking for a
	// token or cookie, or that certain form fields are set and valid

	// for example, this will check if the request was made by a CMS admin user:
	//if !user.IsValid(req) {
	//	return api.ErrNoAuth
	//}

	// you could then to data validation on the request post form, or do it in
	// the Create method, which is called after BeforeAPICreate

	return nil
}

// AfterAPICreate is called after Create, and is useful for logging or triggering
// notifications, etc. after the data is saved to the database, etc.
// The request has a context containing the databse 'target' affected by the
// request. Ex. Song__pending:3 or Song:8 depending if Song implements api.Trustable
func (s *Author) AfterAPICreate(res http.ResponseWriter, req *http.Request) error {
	addr := req.RemoteAddr
	log.Println("AfterAPICreate: Author sent by:", addr, "titled:", req.PostFormValue("title"))

	return nil
}

// Approve implements editor.Mergeable, which enables content supplied by external
// clients to be approved and thus added to the public content API. Before content
// is approved, it is waiting in the Pending bucket, and can only be approved in
// the CMS if the Mergeable interface is satisfied. If not, you will not see this
// content show up in the CMS.
func (s *Author) Approve(res http.ResponseWriter, req *http.Request) error {
	return nil
}

/*
   NOTICE: if AutoApprove (seen below) is implemented, the Approve method above will have no
   effect, except to add the Public / Pending toggle in the CMS UI. Though, no
   Song content would be in Pending, since all externally submitting Song data
   is immediately approved.
*/

// AutoApprove implements api.Trustable, and will automatically approve content
// that has been submitted by an external client via api.Createable. Be careful
// when using AutoApprove, because content will immediately be available through
// your public content API. If the Trustable interface is satisfied, the AfterApprove
// method is bypassed. The
func (s *Author) AutoApprove(res http.ResponseWriter, req *http.Request) error {
	// Use AutoApprove to check for trust-specific headers or whitelisted IPs,
	// etc. Remember, you will not be able to Approve or Reject content that
	// is auto-approved. You could add a field to Song, i.e.
	// AutoApproved bool `json:auto_approved`
	// and set that data here, as it is called before the content is saved, but
	// after the BeforeSave hook.

	return nil
}

func (s *Author) IndexContent() bool {
	return true
}
