package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/strings"
)

const nsStrings = "strings"

func registerStrings(ts strings.Title) {
	f := func() *TemplateFuncsNamespace {
		ctx := strings.New(ts)

		ns := &TemplateFuncsNamespace{
			Name:    nsStrings,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Chomp,
			[]string{"chomp"},
			[][2]string{
				{`{{ chomp "<p>Blockhead</p>\n" | safeHTML }}`, `<p>Blockhead</p>`},
			},
		)

		ns.AddMethodMapping(ctx.CountRunes,
			[]string{"countrunes"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.RuneCount,
			nil,
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.CountWords,
			[]string{"countwords"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Count,
			nil,
			[][2]string{
				{`{{ "aabab" | strings.Count "a" }}`, `3`},
			},
		)

		ns.AddMethodMapping(ctx.Contains,
			nil,
			[][2]string{
				{`{{ strings.Contains "abc" "b" }}`, `true`},
				{`{{ strings.Contains "abc" "d" }}`, `false`},
			},
		)

		ns.AddMethodMapping(ctx.ContainsAny,
			nil,
			[][2]string{
				{`{{ strings.ContainsAny "abc" "bcd" }}`, `true`},
				{`{{ strings.ContainsAny "abc" "def" }}`, `false`},
			},
		)

		ns.AddMethodMapping(ctx.FindRE,
			[]string{"findRE"},
			[][2]string{
				{
					`{{ findRE "[G|g]o" "Hugo is a static side generator written in Go." 1 }}`,
					`[go]`,
				},
			},
		)

		ns.AddMethodMapping(ctx.FindRESubmatch,
			[]string{"findRESubmatch"},
			[][2]string{
				{
					`{{ findRESubmatch §§<a\s*href="(.+?)">(.+?)</a>§§ §§<li><a href="#foo">Foo</a></li> <li><a href="#bar">Bar</a></li>§§ | print | safeHTML }}`,
					"[[<a href=\"#foo\">Foo</a> #foo Foo] [<a href=\"#bar\">Bar</a> #bar Bar]]",
				},
			},
		)

		ns.AddMethodMapping(ctx.HasPrefix,
			[]string{"hasPrefix"},
			[][2]string{
				{`{{ hasPrefix "Hugo" "Hu" }}`, `true`},
				{`{{ hasPrefix "Hugo" "Fu" }}`, `false`},
			},
		)

		ns.AddMethodMapping(ctx.HasSuffix,
			[]string{"hasSuffix"},
			[][2]string{
				{`{{ hasSuffix "Hugo" "go" }}`, `true`},
				{`{{ hasSuffix "Hugo" "du" }}`, `false`},
			},
		)

		ns.AddMethodMapping(ctx.ToLower,
			[]string{"lower"},
			[][2]string{
				{`{{ lower "BatMan" }}`, `batman`},
			},
		)

		ns.AddMethodMapping(ctx.Replace,
			[]string{"replace"},
			[][2]string{
				{
					`{{ replace "Batman and Robin" "Robin" "Catwoman" }}`,
					`Batman and Catwoman`,
				},
				{
					`{{ replace "aabbaabb" "a" "z" 2 }}`,
					`zzbbaabb`,
				},
			},
		)

		ns.AddMethodMapping(ctx.ReplaceRE,
			[]string{"replaceRE"},
			[][2]string{
				{
					`{{ replaceRE "a+b" "X" "aabbaabbab" }}`,
					`XbXbX`,
				},
				{
					`{{ replaceRE "a+b" "X" "aabbaabbab" 1 }}`,
					`Xbaabbab`,
				},
			},
		)

		ns.AddMethodMapping(ctx.SliceString,
			[]string{"slicestr"},
			[][2]string{
				{`{{ slicestr "BatMan" 0 3 }}`, `Bat`},
				{`{{ slicestr "BatMan" 3 }}`, `Man`},
			},
		)

		ns.AddMethodMapping(ctx.Split,
			[]string{"split"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Substr,
			[]string{"substr"},
			[][2]string{
				{`{{ substr "BatMan" 0 -3 }}`, `Bat`},
				{`{{ substr "BatMan" 3 3 }}`, `Man`},
			},
		)

		ns.AddMethodMapping(ctx.Trim,
			[]string{"trim"},
			[][2]string{
				{`{{ trim "++Batman--" "+-" }}`, `Batman`},
			},
		)

		ns.AddMethodMapping(ctx.TrimLeft,
			nil,
			[][2]string{
				{`{{ "aabbaa" | strings.TrimLeft "a" }}`, `bbaa`},
			},
		)

		ns.AddMethodMapping(ctx.TrimSpace,
			nil,
			[][2]string{
				{`{{ " aabbaa" | strings.TrimSpace }}`, `aabbaa`},
			},
		)

		ns.AddMethodMapping(ctx.TrimPrefix,
			nil,
			[][2]string{
				{`{{ "aabbaa" | strings.TrimPrefix "a" }}`, `abbaa`},
				{`{{ "aabbaa" | strings.TrimPrefix "aa" }}`, `bbaa`},
			},
		)

		ns.AddMethodMapping(ctx.TrimRight,
			nil,
			[][2]string{
				{`{{ "aabbaa" | strings.TrimRight "a" }}`, `aabb`},
			},
		)

		ns.AddMethodMapping(ctx.TrimSuffix,
			nil,
			[][2]string{
				{`{{ "aabbaa" | strings.TrimSuffix "a" }}`, `aabba`},
				{`{{ "aabbaa" | strings.TrimSuffix "aa" }}`, `aabb`},
			},
		)

		ns.AddMethodMapping(ctx.Title,
			[]string{"title"},
			[][2]string{
				{`{{ title "Bat man" }}`, `Bat Man`},
				{`{{ title "somewhere over the rainbow" }}`, `Somewhere Over the Rainbow`},
			},
		)

		ns.AddMethodMapping(ctx.FirstUpper,
			nil,
			[][2]string{
				{`{{ "hugo rocks!" | strings.FirstUpper }}`, `Hugo rocks!`},
			},
		)

		ns.AddMethodMapping(ctx.Truncate,
			[]string{"truncate"},
			[][2]string{
				{`{{ "this is a very long text" | truncate 10 " ..." }}`, `this is a ...`},
				{`{{ "With [Markdown](/markdown) inside." | markdownify | truncate 14 }}`, `With <a href="/markdown">Markdown …</a>`},
			},
		)

		ns.AddMethodMapping(ctx.Repeat,
			nil,
			[][2]string{
				{`{{ "yo" | strings.Repeat 4 }}`, `yoyoyoyo`},
			},
		)

		ns.AddMethodMapping(ctx.ToUpper,
			[]string{"upper"},
			[][2]string{
				{`{{ upper "BatMan" }}`, `BATMAN`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
