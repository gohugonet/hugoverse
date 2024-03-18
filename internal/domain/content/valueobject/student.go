package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"net/http"
)

type Student struct {
	Item

	Name string `json:"name"`
}

// MarshalEditor writes a buffer of html to edit a Artist within the CMS
// and implements editor.Editable
func (a *Student) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(a,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Artist field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Name", a, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Namespace here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Artist editor view: %s", err.Error())
	}

	return view, nil
}

// String defines how a Artist is printed. Update it using more descriptive
// fields from the Artist struct type
func (a *Student) String() string {
	return a.Name
}

func (a *Student) Create(res http.ResponseWriter, req *http.Request) error {
	return nil
}

func (a *Student) Approve(http.ResponseWriter, *http.Request) error {
	return nil
}
