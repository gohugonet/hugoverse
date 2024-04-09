package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"strings"
)

type Namespace struct {
	*valueobject.StateMap
}

func (t *Namespace) addTemplate(name string, state *valueobject.State) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	t.Templates[name] = state
}

func (t *Namespace) findTemplate(name string) (*valueobject.State, bool) {
	t.Mu.RLock()
	defer t.Mu.RUnlock()
	state, found := t.Templates[name]
	return state, found
}

func (t *Namespace) newTemplateLookup(in *valueobject.State) func(name string) *valueobject.State {
	return func(name string) *valueobject.State {
		if templ, found := t.Templates[name]; found {
			if templ.IsText() != in.IsText() {
				return nil
			}
			return templ
		}
		if templ, found := findTemplateIn(name, in); found {
			return valueobject.NewTemplateState(templ, valueobject.TemplateInfo{Name: templ.Name()}, nil)
		}
		return nil
	}
}

func findTemplateIn(name string, in template.Preparer) (template.Preparer, bool) {
	in = unwrap(in)
	if text, ok := in.(*texttemplate.Template); ok {
		if templ := text.Lookup(name); templ != nil {
			return templ, true
		}
		return nil, false
	}
	if templ := in.(*htmltemplate.Template).Lookup(name); templ != nil {
		return templ, true
	}
	return nil, false
}

func (t *Namespace) Lookup(name string) (template.Preparer, bool) {
	t.Mu.RLock()
	defer t.Mu.RUnlock()

	templ, found := t.Templates[name]
	if !found {
		return nil, false
	}

	return templ, found
}

func (t *Namespace) getUnregisteredPartials(templ template.Preparer) []*valueobject.State {
	var partials []*valueobject.State

	templs := templates(templ)
	for _, tmpl := range templs {
		if tmpl.Name() == "" || !strings.HasPrefix(tmpl.Name(), "partials/") {
			continue
		}

		_, found := t.findTemplate(tmpl.Name())
		if !found {
			ts := valueobject.NewTemplateState(tmpl, valueobject.TemplateInfo{Name: tmpl.Name()}, nil)
			ts.Typ = template.TypePartial

			partials = append(partials, ts)
		}
	}

	return partials
}

func templates(in template.Preparer) []template.Preparer {
	var templs []template.Preparer
	in = unwrap(in)
	if textt, ok := in.(*texttemplate.Template); ok {
		for _, t := range textt.Templates() {
			templs = append(templs, t)
		}
	}

	if htmlt, ok := in.(*htmltemplate.Template); ok {
		for _, t := range htmlt.Templates() {
			templs = append(templs, t)
		}
	}

	return templs
}
