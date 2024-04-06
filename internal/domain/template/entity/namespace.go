package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"sync"
)

type Namespace struct {
	PrototypeText *texttemplate.Template
	PrototypeHTML *htmltemplate.Template

	readyInit          sync.Once
	prototypeTextClone *texttemplate.Template
	prototypeHTMLClone *htmltemplate.Template

	*valueobject.StateMap
}

func (t *Namespace) parse(info valueobject.Info) (*valueobject.State, error) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	if info.IsText {
		panic("implement me")
	}

	prototype := t.PrototypeHTML

	templ, err := prototype.New(info.Name).Parse(info.Template)
	if err != nil {
		return nil, err
	}

	ts := newTemplateState(templ, info, nil)

	t.Templates[info.Name] = ts

	return ts, nil
}

func newTemplateState(templ template.Preparer, info valueobject.Info, id template.Identity) *valueobject.State {
	if id == nil {
		id = info
	}

	return &valueobject.State{
		Info:     info,
		Typ:      info.ResolveType(),
		Preparer: templ,
		PInfo:    valueobject.DefaultParseInfo,
		Id:       id,
	}
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
			return newTemplateState(templ, valueobject.Info{Name: templ.Name()}, nil)
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

func unwrap(templ template.Preparer) template.Preparer {
	if ts, ok := templ.(*valueobject.State); ok {
		return ts.Preparer
	}
	return templ
}

func (t *Namespace) add(tmpl template.Preparer, state *valueobject.State) error {
	t.Mu.RLock()
	_, found := t.Templates[tmpl.Name()]
	t.Mu.RUnlock()

	if !found {
		t.Mu.Lock()
		// This is a template defined inline.
		_, err := valueobject.ApplyTemplateTransformers(state, t.newTemplateLookup(state))
		if err != nil {
			t.Mu.Unlock()
			return err
		}
		t.Templates[tmpl.Name()] = state
		t.Mu.Unlock()
	}

	return nil
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

func (t *Namespace) MarkReady() error {
	var err error
	t.readyInit.Do(func() {
		// We only need the clones if base templates are in use.
		err = t.createPrototypes()
	})

	return err
}

func (t *Namespace) createPrototypes() error {
	t.prototypeTextClone = texttemplate.Must(t.PrototypeText.Clone())
	t.prototypeHTMLClone = htmltemplate.Must(t.PrototypeHTML.Clone())

	return nil
}
