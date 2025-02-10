package valueobject

import (
	"github.com/disintegration/gift"
	pio "github.com/mdfriday/hugoverse/pkg/io"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"io"
	"strings"
)

var _ gift.Filter = (*textFilter)(nil)

type textFilter struct {
	text        string
	color       color.Color
	x, y        int
	size        float64
	linespacing int
	fontSource  pio.ReadSeekCloserProvider
}

func (f textFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	// Load and parse font
	ttf := goregular.TTF
	if f.fontSource != nil {
		rs, err := f.fontSource.ReadSeekCloser()
		if err != nil {
			panic(err)
		}
		defer rs.Close()
		ttf, err = io.ReadAll(rs)
		if err != nil {
			panic(err)
		}
	}
	otf, err := opentype.Parse(ttf)
	if err != nil {
		panic(err)
	}

	// Set font options
	face, err := opentype.NewFace(otf, &opentype.FaceOptions{
		Size:    f.size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		panic(err)
	}

	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(f.color),
		Face: face,
	}

	gift.New().Draw(dst, src)

	// Draw text, consider and include linebreaks
	maxWidth := dst.Bounds().Dx() - 20
	fontHeight := face.Metrics().Ascent.Ceil()

	// Correct y position based on font and size
	f.y = f.y + fontHeight

	// Start position
	y := f.y
	d.Dot = fixed.P(f.x, f.y)

	// Draw text line by line, breaking each line at the maximum width.
	f.text = strings.ReplaceAll(f.text, "\r", "")
	for _, line := range strings.Split(f.text, "\n") {
		for _, str := range strings.Fields(line) {
			strWidth := font.MeasureString(face, str)
			if (d.Dot.X.Ceil() + strWidth.Ceil()) >= maxWidth {
				y = y + fontHeight + f.linespacing
				d.Dot = fixed.P(f.x, y)
			}
			d.DrawString(str + " ")
		}
		y = y + fontHeight + f.linespacing
		d.Dot = fixed.P(f.x, y)
	}
}

func (f textFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, srcBounds.Dx(), srcBounds.Dy())
}
