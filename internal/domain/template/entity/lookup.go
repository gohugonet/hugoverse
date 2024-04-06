package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
)

type Lookup struct {
	Main *Namespace

	Baseof      map[string]valueobject.Info
	NeedsBaseof map[string]valueobject.Info
}

func (t *Lookup) lookupLayout(d template.LayoutDescriptor) (template.Preparer, bool, error) {
	for _, name := range d.Names() {
		templ, found := t.Main.Lookup(name)
		if found {
			return templ, true, nil
		}
	}
	return nil, false, nil
}

func (t *Lookup) findLayout(d template.LayoutDescriptor) (*valueobject.State, bool, error) {
	for _, name := range d.Names() {
		overlay, found := t.NeedsBaseof[name]

		if !found {
			continue
		}

		var base valueobject.Info
		found = false
		for _, l := range d.BaseNames() {
			base, found = t.Baseof[l]
			if found {
				break
			}
		}

		templ, err := t.applyBaseTemplate(overlay, base)
		if err != nil {
			return nil, false, err
		}

		ts := newTemplateState(templ, overlay, valueobject.IdentityOr(base, overlay))

		if found {
			ts.BaseInfo = base
		}

		return ts, true, nil

	}

	return nil, false, nil
}

func (t *Lookup) applyBaseTemplate(overlay, base valueobject.Info) (template.Preparer, error) {
	if overlay.IsText {
		var (
			templ = t.Main.prototypeTextClone.New(overlay.Name)
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
		templ = t.Main.prototypeHTMLClone.New(overlay.Name)
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
	// * https://github.com/golang/go/issues/16101
	// * https://github.com/gohugoio/hugo/issues/2549
	templ = templ.Lookup(templ.Name())

	return templ, err
}
