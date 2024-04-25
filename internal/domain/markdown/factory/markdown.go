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
		builder := NewGoldMarkBuilder(valueobject.DefaultGoldMarkConf)
		md := builder.Build()

		instance = &entity.Markdown{
			GoldMark:    md,
			Highlighter: valueobject.NewHighlighter(valueobject.DefaultHighlightConfig),
		}
	})

	return instance
}
