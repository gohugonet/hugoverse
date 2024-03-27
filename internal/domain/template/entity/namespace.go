package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/spf13/afero"
	"sync"
)

type TemplateNamespace struct {
	PrototypeText *TextTemplate
	PrototypeHTML *HtmlTemplate

	*TemplateStateMap
}

func (t *TemplateNamespace) Lookup(name string) (template.Template, bool) {
	tmpl, found := t.Templates[name]
	if !found {
		return nil, false
	}

	return tmpl, found
}

func (t *TemplateNamespace) parse(info templateInfo) (*TemplateState, error) {
	prototype := t.PrototypeHTML

	tmpl, err := prototype.New(info.name).Parse(info.template)
	if err != nil {
		return nil, err
	}

	ts := newTemplateState(tmpl, info)

	t.Templates[info.name] = ts

	return ts, nil
}

func newTemplateState(tmpl template.Template, info templateInfo) *TemplateState {
	return &TemplateState{
		info:     info,
		Template: tmpl,
	}
}

type TemplateStateMap struct {
	mu        sync.RWMutex
	Templates map[string]*TemplateState
}

type TemplateState struct {
	template.Template

	parseInfo ParseInfo

	info templateInfo
}

type templateInfo struct {
	name     string
	template string

	// Used to create some error context in error situations
	fs afero.Fs

	// The filename relative to the fs above.
	filename string
}

func (t templateInfo) errWithFileContext(what string, err error) error {
	return fmt.Errorf(what+": %w", err)
}

type ParseInfo struct {
	// Set for shortcode Templates with any {{ .Inner }}
	IsInner bool

	// Set for partials with a return statement.
	HasReturn bool

	// Config extracted from template.
	Config ParseConfig
}

type ParseConfig struct {
	Version int
}
