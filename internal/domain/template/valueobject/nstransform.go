package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/transform"
)

const nsTransform = "transform"

func registerTransform(markdown transform.Markdown) {
	f := func() *TemplateFuncsNamespace {
		ctx := transform.New(markdown)

		ns := &TemplateFuncsNamespace{
			Name:    nsTransform,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Emojify,
			[]string{"emojify"},
			[][2]string{
				{`{{ "I :heart: Hugo" | emojify }}`, `I ❤️ Hugo`},
			},
		)

		ns.AddMethodMapping(ctx.Highlight,
			[]string{"highlight"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.HTMLEscape,
			[]string{"htmlEscape"},
			[][2]string{
				{
					`{{ htmlEscape "Cathal Garvey & The Sunshine Band <cathal@foo.bar>" | safeHTML }}`,
					`Cathal Garvey &amp; The Sunshine Band &lt;cathal@foo.bar&gt;`,
				},
				{
					`{{ htmlEscape "Cathal Garvey & The Sunshine Band <cathal@foo.bar>" }}`,
					`Cathal Garvey &amp;amp; The Sunshine Band &amp;lt;cathal@foo.bar&amp;gt;`,
				},
				{
					`{{ htmlEscape "Cathal Garvey & The Sunshine Band <cathal@foo.bar>" | htmlUnescape | safeHTML }}`,
					`Cathal Garvey & The Sunshine Band <cathal@foo.bar>`,
				},
			},
		)

		ns.AddMethodMapping(ctx.HTMLUnescape,
			[]string{"htmlUnescape"},
			[][2]string{
				{
					`{{ htmlUnescape "Cathal Garvey &amp; The Sunshine Band &lt;cathal@foo.bar&gt;" | safeHTML }}`,
					`Cathal Garvey & The Sunshine Band <cathal@foo.bar>`,
				},
				{
					`{{ "Cathal Garvey &amp;amp; The Sunshine Band &amp;lt;cathal@foo.bar&amp;gt;" | htmlUnescape | htmlUnescape | safeHTML }}`,
					`Cathal Garvey & The Sunshine Band <cathal@foo.bar>`,
				},
				{
					`{{ "Cathal Garvey &amp;amp; The Sunshine Band &amp;lt;cathal@foo.bar&amp;gt;" | htmlUnescape | htmlUnescape }}`,
					`Cathal Garvey &amp; The Sunshine Band &lt;cathal@foo.bar&gt;`,
				},
				{
					`{{ htmlUnescape "Cathal Garvey &amp; The Sunshine Band &lt;cathal@foo.bar&gt;" | htmlEscape | safeHTML }}`,
					`Cathal Garvey &amp; The Sunshine Band &lt;cathal@foo.bar&gt;`,
				},
			},
		)

		ns.AddMethodMapping(ctx.Markdownify,
			[]string{"markdownify"},
			[][2]string{
				{`{{ .Title | markdownify }}`, `<strong>BatMan</strong>`},
			},
		)

		ns.AddMethodMapping(ctx.Plainify,
			[]string{"plainify"},
			[][2]string{
				{`{{ plainify  "Hello <strong>world</strong>, gophers!" }}`, `Hello world, gophers!`},
			},
		)

		ns.AddMethodMapping(ctx.Remarshal,
			nil,
			[][2]string{
				{`{{ "title = \"Hello World\"" | transform.Remarshal "json" | safeHTML }}`, "{\n   \"title\": \"Hello World\"\n}\n"},
			},
		)

		ns.AddMethodMapping(ctx.Unmarshal,
			[]string{"unmarshal"},
			[][2]string{
				{`{{ "hello = \"Hello World\"" | transform.Unmarshal }}`, "map[hello:Hello World]"},
				{`{{ "hello = \"Hello World\"" | resources.FromString "data/greetings.toml" | transform.Unmarshal }}`, "map[hello:Hello World]"},
			},
		)

		ns.AddMethodMapping(ctx.XMLEscape,
			nil,
			[][2]string{
				{
					`{{ transform.XMLEscape "<p>abc</p>" }}`,
					`&lt;p&gt;abc&lt;/p&gt;`,
				},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
