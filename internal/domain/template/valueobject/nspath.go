package valueobject

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/path"
	"path/filepath"
)

const nsPath = "path"

func registerPath() {
	f := func() *TemplateFuncsNamespace {
		ctx := path.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsPath,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Split,
			nil,
			[][2]string{
				{`{{ "/my/path/filename.txt" | path.Split }}`, `/my/path/|filename.txt`},
				{fmt.Sprintf(`{{ %q | path.Split }}`, filepath.FromSlash("/my/path/filename.txt")), `/my/path/|filename.txt`},
			},
		)

		testDir := filepath.Join("my", "path")
		testFile := filepath.Join(testDir, "filename.txt")

		ns.AddMethodMapping(ctx.Join,
			nil,
			[][2]string{
				{fmt.Sprintf(`{{ slice %q "filename.txt" | path.Join }}`, testDir), `my/path/filename.txt`},
				{`{{ path.Join "my" "path" "filename.txt" }}`, `my/path/filename.txt`},
				{fmt.Sprintf(`{{ %q | path.Ext }}`, testFile), `.txt`},
				{fmt.Sprintf(`{{ %q | path.Base }}`, testFile), `filename.txt`},
				{fmt.Sprintf(`{{ %q | path.Dir }}`, testFile), `my/path`},
			},
		)

		return ns
	}
	AddTemplateFuncsNamespace(f)
}
