package admin

import (
	"bytes"
	"github.com/mdfriday/hugoverse/internal/domain/content"
	"html/template"
)

// SetupView ...
func SetupView(name string) ([]byte, error) {
	html := startAdminHTML + initAdminHTML + endAdminHTML

	a := View{
		Logo: name,
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("init").Parse(html))
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type View struct {
	Logo    string
	Types   map[string]content.Creator
	Subview template.HTML
}

func NewView(name string, ts map[string]content.Creator) *View {
	return &View{
		Logo:    name,
		Types:   ts,
		Subview: template.HTML(""),
	}
}

// SubView ...
func (v *View) SubView(view []byte) (_ []byte, err error) {
	//TODO， clean SubView view
	a := View{
		Logo:    v.Logo,
		Types:   v.Types,
		Subview: template.HTML(view),
	}

	buf := &bytes.Buffer{}
	html := startAdminHTML + mainAdminHTML + endAdminHTML
	tmpl := template.Must(template.New("admin").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return
	}

	return buf.Bytes(), nil
}

// Error400 creates a subview for a 400 error page
func (v *View) Error400() ([]byte, error) {
	return v.SubView(err400HTML)
}

// Error404 creates a subview for a 404 error page
func (v *View) Error404() ([]byte, error) {
	return v.SubView(err404HTML)
}

// Error500 creates a subview for a 500 error page
func (v *View) Error500() ([]byte, error) {
	return v.SubView(err500HTML)
}

// Error405 creates a subview for a 405 error page
func (v *View) Error405() ([]byte, error) {
	return v.SubView(err405HTML)
}
