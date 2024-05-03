package images

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"strings"
)

// AddColorToPalette adds c as the first color in p if not already there.
// Note that it does no additional checks, so callers must make sure
// that the palette is valid for the relevant format.
func AddColorToPalette(c color.Color, p color.Palette) color.Palette {
	var found bool
	for _, cc := range p {
		if c == cc {
			found = true
			break
		}
	}

	if !found {
		p = append(color.Palette{c}, p...)
	}

	return p
}

// ReplaceColorInPalette will replace the color in palette p closest to c in Euclidean
// R,G,B,A space with c.
func ReplaceColorInPalette(c color.Color, p color.Palette) {
	p[p.Index(c)] = c
}

// ColorToHexString converts a color to a hex string.
func ColorToHexString(c color.Color) string {
	r, g, b, a := c.RGBA()
	rgba := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	return fmt.Sprintf("#%.2x%.2x%.2x", rgba.R, rgba.G, rgba.B)
}

func HexStringToColor(s string) (color.Color, error) {
	s = strings.TrimPrefix(s, "#")

	if len(s) != 3 && len(s) != 4 && len(s) != 6 && len(s) != 8 {
		return nil, fmt.Errorf("invalid color code: %q", s)
	}

	s = strings.ToLower(s)

	if len(s) == 3 || len(s) == 4 {
		var v string
		for _, r := range s {
			v += string(r) + string(r)
		}
		s = v
	}

	// Standard colors.
	if s == "ffffff" {
		return color.White, nil
	}

	if s == "000000" {
		return color.Black, nil
	}

	// Set Alfa to white.
	if len(s) == 6 {
		s += "ff"
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return color.RGBA{b[0], b[1], b[2], b[3]}, nil
}
