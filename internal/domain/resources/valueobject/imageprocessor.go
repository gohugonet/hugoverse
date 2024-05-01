package valueobject

import (
	"github.com/disintegration/gift"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/image/exif"
	"github.com/muesli/smartcrop"
	"image"
	"image/color"
	"image/draw"
	"io"
)

type ImageProcessor struct {
	ExifDecoder *exif.Decoder
}

func (p *ImageProcessor) DecodeExif(r io.Reader) (*exif.ExifInfo, error) {
	return p.ExifDecoder.Decode(r)
}

func (p *ImageProcessor) FiltersFromConfig(src image.Image, conf ImageConfig) ([]gift.Filter, error) {
	var filters []gift.Filter

	if conf.Rotate != 0 {
		// Apply any rotation before any resize.
		filters = append(filters, gift.Rotate(float32(conf.Rotate), color.Transparent, gift.NearestNeighborInterpolation))
	}

	switch conf.Action {
	case "resize":
		filters = append(filters, gift.Resize(conf.Width, conf.Height, conf.Filter))
	case "crop":
		if conf.AnchorStr == smartCropIdentifier {
			bounds, err := p.smartCrop(src, conf.Width, conf.Height, conf.Filter)
			if err != nil {
				return nil, err
			}

			// First crop using the bounds returned by smartCrop.
			filters = append(filters, gift.Crop(bounds))
			// Then center crop the image to get an image the desired size without resizing.
			filters = append(filters, gift.CropToSize(conf.Width, conf.Height, gift.CenterAnchor))

		} else {
			filters = append(filters, gift.CropToSize(conf.Width, conf.Height, conf.Anchor))
		}
	case "fill":
		if conf.AnchorStr == smartCropIdentifier {
			bounds, err := p.smartCrop(src, conf.Width, conf.Height, conf.Filter)
			if err != nil {
				return nil, err
			}

			// First crop it, then resize it.
			filters = append(filters, gift.Crop(bounds))
			filters = append(filters, gift.Resize(conf.Width, conf.Height, conf.Filter))

		} else {
			filters = append(filters, gift.ResizeToFill(conf.Width, conf.Height, conf.Filter, conf.Anchor))
		}
	case "fit":
		filters = append(filters, gift.ResizeToFit(conf.Width, conf.Height, conf.Filter))
	default:

	}
	return filters, nil
}

func (p *ImageProcessor) ApplyFiltersFromConfig(src image.Image, conf ImageConfig) (image.Image, error) {
	filters, err := p.FiltersFromConfig(src, conf)
	if err != nil {
		return nil, err
	}

	if len(filters) == 0 {
		return p.resolveSrc(src, conf.TargetFormat), nil
	}

	img, err := p.doFilter(src, conf.TargetFormat, filters...)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (p *ImageProcessor) Filter(src image.Image, filters ...gift.Filter) (image.Image, error) {
	return p.doFilter(src, 0, filters...)
}

func (p *ImageProcessor) resolveSrc(src image.Image, targetFormat resources.ImageFormat) image.Image {
	if giph, ok := src.(resources.Giphy); ok {
		g := giph.GIF()
		if len(g.Image) < 2 || (targetFormat == 0 || targetFormat != resources.GIF) {
			src = g.Image[0]
		}
	}
	return src
}

func (p *ImageProcessor) doFilter(src image.Image, targetFormat resources.ImageFormat, filters ...gift.Filter) (image.Image, error) {
	filter := gift.New(filters...)

	if giph, ok := src.(resources.Giphy); ok {
		g := giph.GIF()
		if len(g.Image) < 2 || (targetFormat == 0 || targetFormat != resources.GIF) {
			src = g.Image[0]
		} else {
			var bounds image.Rectangle
			firstFrame := g.Image[0]
			tmp := image.NewNRGBA(firstFrame.Bounds())
			for i := range g.Image {
				gift.New().DrawAt(tmp, g.Image[i], g.Image[i].Bounds().Min, gift.OverOperator)
				bounds = filter.Bounds(tmp.Bounds())
				dst := image.NewPaletted(bounds, g.Image[i].Palette)
				filter.Draw(dst, tmp)
				g.Image[i] = dst
			}
			g.Config.Width = bounds.Dx()
			g.Config.Height = bounds.Dy()

			return giph, nil
		}

	}

	bounds := filter.Bounds(src.Bounds())

	var dst draw.Image
	switch src.(type) {
	case *image.RGBA:
		dst = image.NewRGBA(bounds)
	case *image.NRGBA:
		dst = image.NewNRGBA(bounds)
	case *image.Gray:
		dst = image.NewGray(bounds)
	default:
		dst = image.NewNRGBA(bounds)
	}
	filter.Draw(dst, src)

	return dst, nil
}

func (p *ImageProcessor) smartCrop(img image.Image, width, height int, filter gift.Resampling) (image.Rectangle, error) {
	if width <= 0 || height <= 0 {
		return image.Rectangle{}, nil
	}

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	if srcW <= 0 || srcH <= 0 {
		return image.Rectangle{}, nil
	}

	if srcW == width && srcH == height {
		return srcBounds, nil
	}

	smart := p.newSmartCropAnalyzer(filter)

	rect, err := smart.FindBestCrop(img, width, height)
	if err != nil {
		return image.Rectangle{}, err
	}

	return img.Bounds().Intersect(rect), nil
}

func (p *ImageProcessor) newSmartCropAnalyzer(filter gift.Resampling) smartcrop.Analyzer {
	return smartcrop.NewAnalyzer(imagingResizer{p: p, filter: filter})
}
