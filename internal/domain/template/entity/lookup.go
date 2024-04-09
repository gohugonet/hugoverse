package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
)

type Lookup struct {
	BaseOf *valueobject.BaseOf
}

func (t *Lookup) lookupLayout(d template.LayoutDescriptor, ns *Namespace) (template.Preparer, bool, error) {
	for _, name := range d.Names() {
		templ, found := ns.Lookup(name)
		if found {
			return templ, true, nil
		}
	}
	return nil, false, nil
}

func (t *Lookup) findLayoutInfo(d template.LayoutDescriptor) (valueobject.TemplateInfo, valueobject.TemplateInfo, bool) {
	for _, name := range d.Names() {
		overlay, found := t.BaseOf.GetNeedsBaseOf(name)

		if !found {
			continue
		}

		var base valueobject.TemplateInfo
		found = false
		for _, l := range d.BaseNames() {
			base, found = t.BaseOf.GetBaseOf(l)
			if found {
				break
			}
		}

		return overlay, base, true
	}

	return valueobject.TemplateInfo{}, valueobject.TemplateInfo{}, false
}
