package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/valueobject"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type Builder struct {
	cfg valueobject.GoldMarkConfig
	markdown.Highlighter

	rendererOptions    []renderer.Option
	tocRendererOptions []renderer.Option
	parserOptions      []parser.Option
	extensions         []goldmark.Extender
}

func NewGoldMarkBuilder(cfg valueobject.GoldMarkConfig, highlighter markdown.Highlighter) *Builder {
	return &Builder{
		cfg:         cfg,
		Highlighter: highlighter,

		rendererOptions:    []renderer.Option{},
		tocRendererOptions: nil,
		parserOptions:      []parser.Option{},
		extensions:         []goldmark.Extender{valueobject.NewLinkHooks(cfg.Extensions.LinkifyProtocol)},
	}
}

func (b *Builder) Build() goldmark.Markdown {
	// Render Options
	b.WithHardWraps()
	b.WithXHTML()
	b.WithUnsafe()

	// Extension Options
	b.buildToc()
	b.WithImage()
	b.WithCodeFences()
	b.WithTable()
	b.WithStrikethrough()
	b.WithLinkify()
	b.WithTaskList()
	b.WithDefinitionList()
	b.WithFootnote()
	b.WithTypographer()
	b.WithCJK()
	b.WithPassThrough()
	b.WithEmoji()
	b.WithAttributeBlock()

	// Render Options
	b.WithAutoHeadingID()
	b.WithAttribute()

	return goldmark.New(
		goldmark.WithExtensions(
			b.extensions...,
		),
		goldmark.WithParserOptions(
			b.parserOptions...,
		),
		goldmark.WithRendererOptions(
			b.rendererOptions...,
		),
	)
}

func (b *Builder) buildToc() {
	b.tocRendererOptions = make([]renderer.Option, len(b.rendererOptions))
	if b.rendererOptions != nil {
		copy(b.tocRendererOptions, b.rendererOptions)
	}
	b.WithHTMLRenderer()
	b.WithStrikethroughHTMLRenderer()

	b.extensions = append(b.extensions, valueobject.NewTocExtension(b.tocRendererOptions))
}

func (b *Builder) WithImage() {
	b.extensions = append(b.extensions, valueobject.NewImagesExt(b.cfg.Parser.WrapStandAloneImageWithinParagraph))
}

func (b *Builder) WithCodeFences() {
	if b.cfg.Extensions.Highlight.CodeFences {
		b.extensions = append(b.extensions, valueobject.NewCodeBlocksExt(b.Highlighter))
	}
}

func (b *Builder) WithTable() {
	if b.cfg.Extensions.Table {
		b.extensions = append(b.extensions, extension.Table)
	}
}

func (b *Builder) WithStrikethrough() {
	if b.cfg.Extensions.Strikethrough {
		b.extensions = append(b.extensions, extension.Strikethrough)
	}
}

func (b *Builder) WithLinkify() {
	if b.cfg.Extensions.Linkify {
		b.extensions = append(b.extensions, extension.Linkify)
	}
}

func (b *Builder) WithTaskList() {
	if b.cfg.Extensions.TaskList {
		b.extensions = append(b.extensions, extension.TaskList)
	}
}

func (b *Builder) WithDefinitionList() {
	if b.cfg.Extensions.DefinitionList {
		b.extensions = append(b.extensions, extension.DefinitionList)
	}
}

func (b *Builder) WithFootnote() {
	if b.cfg.Extensions.Footnote {
		b.extensions = append(b.extensions, extension.Footnote)
	}
}

func (b *Builder) WithCJK() {
	if b.cfg.Extensions.CJK.Enable {
		var opts []extension.CJKOption

		if b.cfg.Extensions.CJK.EastAsianLineBreaks {
			if b.cfg.Extensions.CJK.EastAsianLineBreaksStyle == "css3draft" {
				opts = append(opts, extension.WithEastAsianLineBreaks(extension.EastAsianLineBreaksCSS3Draft))
			} else {
				opts = append(opts, extension.WithEastAsianLineBreaks())
			}
		}

		if b.cfg.Extensions.CJK.EscapedSpace {
			opts = append(opts, extension.WithEscapedSpace())
		}
		c := extension.NewCJK(opts...)

		b.extensions = append(b.extensions, c)
	}
}

func (b *Builder) WithPassThrough() {
	if b.cfg.Extensions.Passthrough.Enable {
		configuredInlines := b.cfg.Extensions.Passthrough.Delimiters.Inline
		configuredBlocks := b.cfg.Extensions.Passthrough.Delimiters.Block

		inlineDelimiters := make([]valueobject.PassThroughDelimiters, len(configuredInlines))
		blockDelimiters := make([]valueobject.PassThroughDelimiters, len(configuredBlocks))

		for i, d := range configuredInlines {
			inlineDelimiters[i] = valueobject.PassThroughDelimiters{
				Open:  d[0],
				Close: d[1],
			}
		}

		for i, d := range configuredBlocks {
			blockDelimiters[i] = valueobject.PassThroughDelimiters{
				Open:  d[0],
				Close: d[1],
			}
		}

		b.extensions = append(b.extensions, valueobject.NewPassThroughExt(
			valueobject.PassThroughConfig{
				InlineDelimiters: inlineDelimiters,
				BlockDelimiters:  blockDelimiters,
			},
		))
	}
}

func (b *Builder) WithEmoji() {
	if b.cfg.Extensions.Emoji {
		b.extensions = append(b.extensions, emoji.Emoji)
	}
}

func (b *Builder) WithAttributeBlock() {
	if b.cfg.Parser.Attribute.Block {
		b.extensions = append(b.extensions, valueobject.NewAttrExt())
	}
}

func (b *Builder) WithTypographer() {
	if !b.cfg.Extensions.Typographer.Disable {
		t := extension.NewTypographer(
			extension.WithTypographicSubstitutions(toTypographicPunctuationMap(b.cfg.Extensions.Typographer)),
		)
		b.extensions = append(b.extensions, t)
	}
}

func (b *Builder) WithHTMLRenderer() {
	b.tocRendererOptions = append(b.tocRendererOptions,
		renderer.WithNodeRenderers(util.Prioritized(emoji.NewHTMLRenderer(), 200)))
}

func (b *Builder) WithStrikethroughHTMLRenderer() {
	b.tocRendererOptions = append(b.tocRendererOptions,
		renderer.WithNodeRenderers(util.Prioritized(extension.NewStrikethroughHTMLRenderer(), 500)))
}

func (b *Builder) WithHardWraps() {
	if b.cfg.Renderer.HardWraps {
		b.rendererOptions = append(b.rendererOptions, html.WithHardWraps())
	}
}

func (b *Builder) WithXHTML() {
	if b.cfg.Renderer.XHTML {
		b.rendererOptions = append(b.rendererOptions, html.WithXHTML())
	}
}

func (b *Builder) WithUnsafe() {
	if b.cfg.Renderer.Unsafe {
		b.rendererOptions = append(b.rendererOptions, html.WithUnsafe())
	}
}

func (b *Builder) WithAutoHeadingID() {
	if b.cfg.Parser.AutoHeadingID {
		b.parserOptions = append(b.parserOptions, parser.WithAutoHeadingID())
	}
}

func (b *Builder) WithAttribute() {
	if b.cfg.Parser.Attribute.Title {
		b.parserOptions = append(b.parserOptions, parser.WithAttribute())
	}
}

// Note: It's tempting to put this in the config package, but that doesn't work.
func toTypographicPunctuationMap(t valueobject.Typographer) map[extension.TypographicPunctuation][]byte {
	return map[extension.TypographicPunctuation][]byte{
		extension.LeftSingleQuote:  []byte(t.LeftSingleQuote),
		extension.RightSingleQuote: []byte(t.RightSingleQuote),
		extension.LeftDoubleQuote:  []byte(t.LeftDoubleQuote),
		extension.RightDoubleQuote: []byte(t.RightDoubleQuote),
		extension.EnDash:           []byte(t.EnDash),
		extension.EmDash:           []byte(t.EmDash),
		extension.Ellipsis:         []byte(t.Ellipsis),
		extension.LeftAngleQuote:   []byte(t.LeftAngleQuote),
		extension.RightAngleQuote:  []byte(t.RightAngleQuote),
		extension.Apostrophe:       []byte(t.Apostrophe),
	}
}
