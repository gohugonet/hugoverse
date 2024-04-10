package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/entity"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/valueobject"
)

func NewMarkdown() markdown.Markdown {
	builder := NewGoldMarkBuilder(valueobject.DefaultGoldMarkConf)
	md := builder.Build()

	return &entity.Markdown{GoldMark: md}
}
