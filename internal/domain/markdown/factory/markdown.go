package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/entity"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/valueobject"
	"sync"
)

var instance *entity.Markdown
var once sync.Once

func NewMarkdown() markdown.Markdown {
	once.Do(func() {
		hl := valueobject.NewDefaultHighlighter()

		builder := NewGoldMarkBuilder(valueobject.DefaultGoldMarkConf, hl)
		md := builder.Build()

		instance = &entity.Markdown{
			GoldMark:    md,
			Highlighter: hl,
		}
	})

	return instance
}
