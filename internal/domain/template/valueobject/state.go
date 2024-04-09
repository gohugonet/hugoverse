package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"sync"
)

type StateMap struct {
	Mu        sync.RWMutex
	Templates map[string]*State
}

type State struct {
	template.Preparer

	Typ   template.Type
	PInfo ParseInfo
	Id    template.Identity

	Info     TemplateInfo
	BaseInfo TemplateInfo // Set when a base template is used.
}

func NewTemplateState(templ template.Preparer, info TemplateInfo, id template.Identity) *State {
	if id == nil {
		id = info
	}

	return &State{
		Info:     info,
		Typ:      info.ResolveType(),
		Preparer: templ,
		PInfo:    DefaultParseInfo,
		Id:       id,
	}
}

func (t *State) IsInternalTemplate() bool {
	return t.Info.IsEmbedded
}

func (t *State) GetIdentity() template.Identity {
	return t.Id
}

func (t *State) ParseInfo() ParseInfo {
	return t.PInfo
}

func (t *State) IsText() bool {
	return isText(t.Preparer)
}

func (t *State) String() string {
	return t.Name()
}

func isText(templ template.Preparer) bool {
	_, isText := templ.(*texttemplate.Template)
	return isText
}
