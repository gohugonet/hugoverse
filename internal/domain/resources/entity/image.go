// Copyright 2019 The Hugo Authors. All rights reserved.
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

package entity

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/images/webp"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"sync"

	"github.com/bep/gowebp/libwebp/webpoptions"

	"github.com/disintegration/gift"
)

func NewImage(f resources.ImageFormat, proc *valueobject.ImageProcessor, img image.Image, s Spec, c *ImageCache) *Image {
	if img != nil {
		return &Image{
			ImageFormat: f,
			Proc:        proc,
			Spec:        s,
			imageConfig: &imageConfig{
				config:       imageConfigFromImage(img),
				configLoaded: true,
			},
			ImageCache: c,
		}
	}
	return &Image{
		ImageFormat: f, Proc: proc, Spec: s,
		imageConfig: &imageConfig{}, ImageCache: c}
}

type Image struct {
	ImageFormat resources.ImageFormat
	Proc        *valueobject.ImageProcessor
	Spec        Spec
	*imageConfig
	ImageCache *ImageCache
}

func (i *Image) EncodeTo(conf valueobject.ImageConfig, img image.Image, w io.Writer) error {
	switch conf.TargetFormat {
	case resources.JPEG:

		var rgba *image.RGBA
		quality := conf.Quality

		if nrgba, ok := img.(*image.NRGBA); ok {
			if nrgba.Opaque() {
				rgba = &image.RGBA{
					Pix:    nrgba.Pix,
					Stride: nrgba.Stride,
					Rect:   nrgba.Rect,
				}
			}
		}
		if rgba != nil {
			return jpeg.Encode(w, rgba, &jpeg.Options{Quality: quality})
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case resources.PNG:
		encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
		return encoder.Encode(w, img)

	case resources.GIF:
		if giphy, ok := img.(resources.Giphy); ok {
			g := giphy.GIF()
			return gif.EncodeAll(w, g)
		}
		return gif.Encode(w, img, &gif.Options{
			NumColors: 256,
		})
	case resources.TIFF:
		return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true})

	case resources.BMP:
		return bmp.Encode(w, img)
	case resources.WEBP:
		return webp.Encode(
			w,
			img, webpoptions.EncodingOptions{
				Quality:        conf.Quality,
				EncodingPreset: webpoptions.EncodingPreset(conf.Hint),
				UseSharpYuv:    true,
			},
		)
	default:
		return errors.New("format not supported")
	}
}

// Height returns i's height.
func (i *Image) Height() int {
	i.initConfig()
	return i.config.Height
}

// Width returns i's width.
func (i *Image) Width() int {
	i.initConfig()
	return i.config.Width
}

func (i *Image) WithImage(img image.Image) *Image {
	i.Spec = nil
	i.imageConfig = &imageConfig{
		config:       imageConfigFromImage(img),
		configLoaded: true,
	}

	return i
}

func (i *Image) WithSpec(s Spec) *Image {
	i.Spec = s
	i.imageConfig = &imageConfig{}
	return i
}

// InitConfig reads the images config from the given reader.
func (i *Image) InitConfig(r io.Reader) error {
	var err error
	i.configInit.Do(func() {
		i.config, _, err = image.DecodeConfig(r)
	})
	return err
}

func (i *Image) initConfig() error {
	var err error
	i.configInit.Do(func() {
		if i.configLoaded {
			return
		}

		var f pio.ReadSeekCloser

		f, err = i.Spec.ReadSeekCloser()
		if err != nil {
			return
		}
		defer f.Close()

		i.config, _, err = image.DecodeConfig(f)
	})

	if err != nil {
		return fmt.Errorf("failed to load images config: %w", err)
	}

	return nil
}

func GetDefaultImageConfig(action string, defaults resources.Image) valueobject.ImageConfig {
	return valueobject.ImageConfig{
		Action:  action,
		Hint:    defaults.ImageHint(),
		Quality: defaults.ImageQuality(),
	}
}

type Spec interface {
	// Loads the images source.
	ReadSeekCloser() (pio.ReadSeekCloser, error)
}

type imageConfig struct {
	config       image.Config
	configInit   sync.Once
	configLoaded bool
}

func imageConfigFromImage(img image.Image) image.Config {
	if giphy, ok := img.(resources.Giphy); ok {
		return giphy.GIF().Config
	}
	b := img.Bounds()
	return image.Config{Width: b.Max.X, Height: b.Max.Y}
}

// UnwrapFilter unwraps the given filter if it is a filter wrapper.
func UnwrapFilter(in gift.Filter) gift.Filter {
	if f, ok := in.(valueobject.Filter); ok {
		return f.Filter
	}
	return in
}

// ToFilters converts the given input to a slice of gift.Filter.
func ToFilters(in any) []gift.Filter {
	switch v := in.(type) {
	case []gift.Filter:
		return v
	case []valueobject.Filter:
		vv := make([]gift.Filter, len(v))
		for i, f := range v {
			vv[i] = f
		}
		return vv
	case gift.Filter:
		return []gift.Filter{v}
	default:
		panic(fmt.Sprintf("%T is not an images filter", in))
	}
}

// IsOpaque returns false if the images has alpha channel and there is at least 1
// pixel that is not (fully) opaque.
func IsOpaque(img image.Image) bool {
	if oim, ok := img.(interface {
		Opaque() bool
	}); ok {
		return oim.Opaque()
	}

	return false
}

// ImageSource identifies and decodes an images.
type ImageSource interface {
	DecodeImage() (image.Image, error)
	Key() string
}
