package valueobject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/editor"
	"github.com/gohugonet/hugoverse/pkg/language"
	"net/http"
	"text/template"
)

type Site struct {
	Item

	Title       string   `json:"title"`
	Description string   `json:"description"`
	BaseURL     string   `json:"base_url"`
	Theme       string   `json:"theme"`
	Params      string   `json:"params"`
	Owner       string   `json:"owner"`
	WorkingDir  string   `json:"working_dir"`
	Languages   []string `json:"languages"`
}

// MarshalEditor writes a buffer of html to edit a Song within the CMS
// and implements editor.Editable
func (s *Site) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(s,
		editor.Field{
			View: editor.Input("Title", s, map[string]string{
				"label":       "Title",
				"type":        "text",
				"placeholder": "Enter the title here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Description", s, map[string]string{
				"label":       "Description",
				"type":        "textarea",
				"placeholder": "Enter the description here",
			}),
		},
		editor.Field{
			View: editor.Input("BaseURL", s, map[string]string{
				"label":       "BaseURL",
				"type":        "text",
				"placeholder": "Enter the base URL here",
			}),
		},
		editor.Field{
			View: editor.Input("Theme", s, map[string]string{
				"label":       "Theme",
				"type":        "text",
				"placeholder": "Enter the Theme URL here",
			}),
		},
		editor.Field{
			View: editor.Textarea("Params", s, map[string]string{
				"label":       "Params",
				"type":        "textarea",
				"placeholder": "Enter the params here in yaml format",
			}),
		},
		editor.Field{
			View: editor.Input("Owner", s, map[string]string{
				"label":       "Owner",
				"type":        "text",
				"placeholder": "Enter the owner user email here",
			}),
		},
		editor.Field{
			View: editor.Input("WorkingDir", s, map[string]string{
				"label":       "WorkingDir",
				"type":        "text",
				"placeholder": "Enter the project file system dir here",
			}),
		},
		editor.Field{
			View: editor.Input("Languages", s, map[string]string{
				"label":       "Languages",
				"type":        "text",
				"placeholder": "Enter the Languages here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to render Site editor view: %s", err.Error())
	}

	return view, nil
}

// String defines the display name of a Song in the CMS list-view
func (s *Site) String() string { return s.Title }

// Create implements api.Createable, and allows external POST requests from clients
// to add content as long as the request contains the json tag names of the Song
// struct fields, and is multipart encoded
func (s *Site) Create(res http.ResponseWriter, req *http.Request) error {
	// do form data validation for required fields
	required := []string{
		"title",
		"theme",
		"owner",
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
func (s *Site) BeforeAPICreate(res http.ResponseWriter, req *http.Request) error {
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
func (s *Site) AfterAPICreate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// Approve implements editor.Mergeable, which enables content supplied by external
// clients to be approved and thus added to the public content API. Before content
// is approved, it is waiting in the Pending bucket, and can only be approved in
// the CMS if the Mergeable interface is satisfied. If not, you will not see this
// content show up in the CMS.
func (s *Site) Approve(res http.ResponseWriter, req *http.Request) error {
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
func (s *Site) AutoApprove(res http.ResponseWriter, req *http.Request) error {
	// Use AutoApprove to check for trust-specific headers or whitelisted IPs,
	// etc. Remember, you will not be able to Approve or Reject content that
	// is auto-approved. You could add a field to Song, i.e.
	// AutoApproved bool `json:auto_approved`
	// and set that data here, as it is called before the content is saved, but
	// after the BeforeSave hook.

	return nil
}

func (s *Site) IndexContent() bool {
	return true
}

func (s *Site) Build() bool {
	return true
}

func (s *Site) Deploy() bool {
	return true
}

func (s *Site) Toml() ([]byte, error) {
	const tomlTemplate = `title = "{{.Title}}"
description = "{{.Description}}"
baseURL = "{{.BaseURL}}"
owner = "{{.Owner}}"

[module]
  [[module.imports]]
    path = "{{.Theme}}"

{{- if .IsMultiLanguages}}
[languages]
{{- range $index, $lang := .Languages }}
  [languages.{{ $lang }}]
  languageName = '{{ getLanguageName $lang }}'
  contentDir = 'content.{{ $lang }}'
  weight = {{ add $index 1 }}
{{- end }}
{{- end }}

[params]
{{.Params}}

`
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"getLanguageName": language.GetLanguageName,
	}

	tmpl, err := template.New("toml").Funcs(funcMap).Parse(tomlTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse toml template error : %v", err)
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, s); err != nil {
		return nil, fmt.Errorf("execute toml template error: %v", err)
	}

	return result.Bytes(), nil
}

func (s *Site) IsMultiLanguages() bool {
	return len(s.Languages) > 1
}

func (s *Site) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with the same fields
	type Alias Site
	temp := &struct {
		Languages interface{} `json:"languages"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	// Unmarshal the JSON into the temp struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Handle the "languages" field
	switch v := temp.Languages.(type) {
	case nil:
		// If it's nil or an empty string, set Languages as an empty array
		s.Languages = []string{}
	case string:
		// If it's a single string, wrap it in an array
		s.Languages = []string{v}
	case []interface{}:
		// If it's an array, convert it into a slice of strings
		for _, item := range v {
			if str, ok := item.(string); ok {
				s.Languages = append(s.Languages, str)
			}
		}
	}

	return nil
}
