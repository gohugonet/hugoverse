package template

import (
	"context"
	template "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"github.com/spf13/afero"
	"io"
)

type Type int

const (
	TypeUndefined Type = iota
	TypeShortcode
	TypePartial
)

type Fs interface {
	LayoutFs() afero.Fs
}

type Template interface {
	Executor
	Lookup
}

type Executor interface {
	ExecuteWithContext(ctx context.Context, t Preparer, wr io.Writer, data any) error
}

type Lookup interface {
	LookupLayout(d LayoutDescriptor) (Preparer, bool, error)
}

type LayoutDescriptor interface {
	Names() []string
	BaseNames() []string
}

type Preparer interface {
	Name() string
	template.Preparer
}

type Identity interface {
	IdentifierBase() string
}
