package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
)

type Lang struct {
	sourceLangMap map[string]string
}

func NewLang(ms []module.Module) *Lang {
	l := &Lang{
		sourceLangMap: make(map[string]string),
	}
	for _, m := range ms {
		for _, mount := range m.Mounts() {
			l.sourceLangMap[mount.Source()] = mount.Lang()
		}
	}
	return l
}

func (l *Lang) GetSourceLang(source string) (string, bool) {
	lang, ok := l.sourceLangMap[source]
	return lang, ok
}
