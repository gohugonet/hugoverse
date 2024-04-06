package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"io"
)

type Executor struct {
	texttemplate.Executor
}

func (t *Executor) ExecuteWithContext(ctx context.Context, templ template.Preparer, wr io.Writer, data any) error {
	return t.Executor.ExecuteWithContext(ctx, templ, wr, data)
}
