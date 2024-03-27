package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	pkgTmpl "github.com/gohugonet/hugoverse/pkg/template"
)

// HtmlTemplate is a specialized ExecTemplate from "text/template" that produces a safe
// HTML document fragment.
type HtmlTemplate struct {
	// We could embed the text/template field, but it's safer not to because
	// we need to keep our version of the name space and the underlying
	// template's in sync.
	Text *TextTemplate

	*NameSpace // common to all associated templates
}

// NameSpace is the data structure shared by all templates in an association.
type NameSpace struct {
	Set map[string]*HtmlTemplate
}

// Funcs adds the elements of the argument map to the template's function map.
// It must be called before the template is parsed.
// It panics if a value in the map is not a function with appropriate return
// type. However, it is legal to overwrite elements of the map. The return
// value is the template, so calls can be chained.
func (t *HtmlTemplate) Funcs(funcMap template.FuncMap) *HtmlTemplate {
	t.Text.Funcs(funcMap)
	return t
}

// New allocates a new, undefined template associated with the given one and with the same
// delimiters. The association, which is transitive, allows one template to
// invoke another with a {{template}} action.
//
// Because associated templates share underlying data, template construction
// cannot be done safely in parallel. Once the templates are constructed, they
// can be executed in parallel.
func (t *HtmlTemplate) New(name string) *HtmlTemplate {
	return &HtmlTemplate{
		Text:      t.Text.New(name),
		NameSpace: t.NameSpace,
	}
}

func (t *HtmlTemplate) Parse(text string) (*HtmlTemplate, error) {
	_, err := t.Text.Parse(text)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *HtmlTemplate) Name() string {
	return t.Text.Name
}

// Prepare returns a template ready for execution.
func (t *HtmlTemplate) Prepare() (template.ExecTemplate, error) {
	doc, err := pkgTmpl.Escape(t.Text.Doc)
	if err != nil {
		return nil, err
	}
	return &ExecTemplate{
		name: t.Name(),
		doc:  doc,
	}, nil
}
