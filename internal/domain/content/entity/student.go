package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/editor"
)

type Student struct {
	Item

	Name string `json:"name"`
	Song string `json:"song"`
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
				"placeholder": "Enter the Name here",
			}),
		},
		editor.Field{
			View: editor.RefSelect("Song", a, map[string]string{
				"label": "Song",
			},
				"Song",
				`{{ .title }} `,
			),
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
	return fmt.Sprintf("Artist: %s", a.UUID)
}
