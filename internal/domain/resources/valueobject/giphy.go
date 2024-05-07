package valueobject

import (
	"image"
	"image/gif"
)

type Giphy struct {
	image.Image
	Gif *gif.GIF
}

func (g *Giphy) GIF() *gif.GIF {
	return g.Gif
}
