// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/partials"
)

const nsPartials = "partials"

func registerPartials(cb func(ctx context.Context, name string, data any) (tmpl, res string, err error)) {
	f := func() *TemplateFuncsNamespace {
		ctx := partials.New(cb)

		ns := &TemplateFuncsNamespace{
			Name:    nsPartials,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Include,
			[]string{"partial"},
			[][2]string{
				{`{{ partial "header.html" . }}`, `<title>Hugo Rocks!</title>`},
			},
		)

		// TODO(bep) we need the return to be a valid identifier, but
		// should consider another way of adding it.
		ns.AddMethodMapping(func() string { return "" },
			[]string{"return"},
			[][2]string{},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
