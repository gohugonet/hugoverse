package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"sync"
)

type Parser struct {
	PrototypeText *texttemplate.Template
	PrototypeHTML *htmltemplate.Template

	readyInit          sync.Once
	prototypeTextClone *texttemplate.Template
	prototypeHTMLClone *htmltemplate.Template

	Ast *AstTransformer

	*sync.RWMutex
}

func (t *Parser) ParseWithLock(name, tpl string) (template.Preparer, error) {
	t.Lock()
	defer t.Unlock()
	return t.PrototypeText.New(name).Parse(tpl)
}

func (t *Parser) ParseOverlap(overlay, base valueobject.TemplateInfo) (*valueobject.State, bool, error) {
	templ, err := t.applyBaseTemplate(overlay, base)
	if err != nil {
		return nil, false, err
	}

	ts := valueobject.NewTemplateState(templ, overlay, valueobject.IdentityOr(base, overlay))

	if !base.IsZero() {
		ts.BaseInfo = base
	}

	return ts, true, nil
}

func (t *Parser) applyBaseTemplate(overlay, base valueobject.TemplateInfo) (template.Preparer, error) {
	if overlay.IsText {
		var (
			templ = t.prototypeTextClone.New(overlay.Name)
			err   error
		)

		if !base.IsZero() {
			templ, err = templ.Parse(base.Template)
			if err != nil {
				return nil, base.ErrWithFileContext("parse failed", err)
			}
		}

		templ, err = texttemplate.Must(templ.Clone()).Parse(overlay.Template)
		if err != nil {
			return nil, overlay.ErrWithFileContext("parse failed", err)
		}

		return templ, nil
	}

	var (
		templ = t.prototypeHTMLClone.New(overlay.Name)
		err   error
	)

	if !base.IsZero() {
		templ, err = templ.Parse(base.Template)
		if err != nil {
			return nil, base.ErrWithFileContext("parse failed", err)
		}
	}

	templ, err = htmltemplate.Must(templ.Clone()).Parse(overlay.Template)
	if err != nil {
		return nil, overlay.ErrWithFileContext("parse failed", err)
	}

	// The extra lookup is a workaround, see
	templ = templ.Lookup(templ.Name())

	return templ, err
}

func (t *Parser) Parse(info valueobject.TemplateInfo) (*valueobject.State, error) {
	if info.IsText {
		return t.parseText(info)
	}

	return t.parseHtml(info)
}

func (t *Parser) parseText(info valueobject.TemplateInfo) (*valueobject.State, error) {
	prototype := t.PrototypeText

	templ, err := prototype.New(info.Name).Parse(info.Template)
	if err != nil {
		return nil, err
	}

	ts := valueobject.NewTemplateState(templ, info, nil)

	return ts, nil
}

func (t *Parser) parseHtml(info valueobject.TemplateInfo) (*valueobject.State, error) {
	prototype := t.PrototypeHTML

	templ, err := prototype.New(info.Name).Parse(info.Template)
	if err != nil {
		return nil, err
	}

	ts := valueobject.NewTemplateState(templ, info, nil)

	return ts, nil
}

func (t *Parser) MarkReady() error {
	var err error
	t.readyInit.Do(func() {
		// We only need the clones if base templates are in use.
		err = t.createPrototypes()
	})

	return err
}

func (t *Parser) createPrototypes() error {
	t.prototypeTextClone = texttemplate.Must(t.PrototypeText.Clone())
	t.prototypeHTMLClone = htmltemplate.Must(t.PrototypeHTML.Clone())

	return nil
}

func (t *Parser) Transform(ns *Namespace, ts *valueobject.State) error {
	_, err := t.Ast.applyTemplateTransformers(ns.newTemplateLookup, ts)
	if err != nil {
		return err
	}
	return nil
}
