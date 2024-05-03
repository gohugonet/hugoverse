package entity

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/hstrings"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/images"
	"github.com/gohugonet/hugoverse/pkg/images/exif"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/png"
	"io"
	"os"
	"strings"
	"sync"

	color_extractor "github.com/marekm4/color-extractor"

	"github.com/disintegration/gift"

	// Blind import for images.Decode
	_ "golang.org/x/image/webp"
)

var (
	_ resources.ImageResource          = (*imageResource)(nil)
	_ resources.Source                 = (*imageResource)(nil)
	_ resources.Cloner                 = (*imageResource)(nil)
	_ resources.NameNormalizedProvider = (*imageResource)(nil)
)

// imageResource represents an images resource.
type imageResource struct {
	*valueobject.Image

	ImageService resources.Image // TODO to be assigned

	ImageCache *ImageCache

	// When a images is processed in a chain, this holds the reference to the
	// original (first).
	root *imageResource

	metaInit    sync.Once
	metaInitErr error
	meta        *imageMeta

	dominantColorInit sync.Once
	dominantColors    []string

	baseResource
}

type imageMeta struct {
	Exif *exif.ExifInfo
}

func (i *imageResource) Exif() *exif.ExifInfo {
	return i.root.getExif()
}

func (i *imageResource) getExif() *exif.ExifInfo {
	i.metaInit.Do(func() {
		supportsExif := i.ImageFormat == resources.JPEG || i.ImageFormat == resources.TIFF
		if !supportsExif {
			return
		}

		key := i.getImageMetaCacheTargetPath()

		read := func(info filecache.ItemInfo, r io.ReadSeeker) error {
			meta := &imageMeta{}
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			if err = json.Unmarshal(data, &meta); err != nil {
				return err
			}

			i.meta = meta

			return nil
		}

		create := func(info filecache.ItemInfo, w io.WriteCloser) (err error) {
			defer w.Close()
			f, err := i.root.ReadSeekCloser()
			if err != nil {
				i.metaInitErr = err
				return
			}
			defer f.Close()

			x, err := i.Image.Proc.DecodeExif(f)
			if err != nil {
				fmt.Printf("Unable to decode Exif metadata from images: %s", i.Key())
				return nil
			}

			i.meta = &imageMeta{Exif: x}

			// Also write it to cache
			enc := json.NewEncoder(w)
			return enc.Encode(i.meta)
		}

		_, i.metaInitErr = i.ImageCache.Fcache.ReadOrCreate(key, read, create)
	})

	if i.metaInitErr != nil {
		panic(fmt.Sprintf("metadata init failed: %s", i.metaInitErr))
	}

	if i.meta == nil {
		return nil
	}

	return i.meta.Exif
}

// Colors returns a slice of the most dominant colors in an images
// using a simple histogram method.
func (i *imageResource) Colors() ([]string, error) {
	var err error
	i.dominantColorInit.Do(func() {
		var img image.Image
		img, err = i.DecodeImage()
		if err != nil {
			return
		}
		colors := color_extractor.ExtractColors(img)
		for _, c := range colors {
			i.dominantColors = append(i.dominantColors, images.ColorToHexString(c))
		}
	})
	return i.dominantColors, nil
}

// Clone is for internal use.
func (i *imageResource) Clone() resources.Resource {
	gr := i.baseResource.Clone().(baseResource)
	return &imageResource{
		root:         i.root,
		Image:        i.WithSpec(gr),
		baseResource: gr,
	}
}

func (i *imageResource) cloneTo(targetPath string) resources.Resource {
	gr := i.baseResource.CloneTo(targetPath).(baseResource)
	return &imageResource{
		root:         i.root,
		Image:        i.WithSpec(gr),
		baseResource: gr,
	}
}

func (i *imageResource) cloneWithUpdates(u *valueobject.TransformationUpdate) (baseResource, error) {
	base, err := i.baseResource.cloneWithUpdates(u) //TODO
	if err != nil {
		return nil, err
	}

	var img *valueobject.Image

	if u.IsContentChanged() {
		img = i.WithSpec(base)
	} else {
		img = i.Image
	}

	return &imageResource{
		root:         i.root,
		Image:        img,
		baseResource: base,
	}, nil
}

var imageActions = []string{valueobject.ActionResize, valueobject.ActionCrop, valueobject.ActionFit, valueobject.ActionFill}

// Process processes the images with the given spec.
// The spec can contain an optional action, one of "resize", "crop", "fit" or "fill".
// This makes this method a more flexible version that covers all of Resize, Crop, Fit and Fill,
// but it also supports e.g. format conversions without any resize action.
func (i *imageResource) Process(spec string) (resources.ImageResource, error) {
	action, options := i.resolveActionOptions(spec)
	return i.processActionOptions(action, options)
}

// Resize resizes the images to the specified width and height using the specified resampling
// filter and returns the transformed images. If one of width or height is 0, the images aspect
// ratio is preserved.
func (i *imageResource) Resize(spec string) (resources.ImageResource, error) {
	return i.processActionSpec(valueobject.ActionResize, spec)
}

// Crop the images to the specified dimensions without resizing using the given anchor point.
// Space delimited config, e.g. `200x300 TopLeft`.
func (i *imageResource) Crop(spec string) (resources.ImageResource, error) {
	return i.processActionSpec(valueobject.ActionCrop, spec)
}

// Fit scales down the images using the specified resample filter to fit the specified
// maximum width and height.
func (i *imageResource) Fit(spec string) (resources.ImageResource, error) {
	return i.processActionSpec(valueobject.ActionFit, spec)
}

// Fill scales the images to the smallest possible size that will cover the specified dimensions,
// crops the resized images to the specified dimensions using the given anchor point.
// Space delimited config, e.g. `200x300 TopLeft`.
func (i *imageResource) Fill(spec string) (resources.ImageResource, error) {
	return i.processActionSpec(valueobject.ActionFill, spec)
}

func (i *imageResource) Filter(filters ...any) (resources.ImageResource, error) {
	var conf valueobject.ImageConfig

	var gfilters []gift.Filter

	for _, f := range filters {
		gfilters = append(gfilters, valueobject.ToFilters(f)...)
	}

	var (
		targetFormat resources.ImageFormat
		configSet    bool
	)
	for _, f := range gfilters {
		f = valueobject.UnwrapFilter(f)
		if specProvider, ok := f.(valueobject.ImageProcessSpecProvider); ok {
			action, options := i.resolveActionOptions(specProvider.ImageProcessSpec())
			var err error
			conf, err = valueobject.DecodeImageConfig(action, options, i.ImageService, i.ImageFormat)
			if err != nil {
				return nil, err
			}
			configSet = true
			if conf.TargetFormat != 0 {
				targetFormat = conf.TargetFormat
				// We only support one target format, but prefer the last one,
				// so we keep going.
			}
		}
	}

	if !configSet {
		conf = valueobject.GetDefaultImageConfig("filter", i.ImageService)
	}

	conf.Action = "filter"
	conf.Key = identity.HashString(gfilters)
	conf.TargetFormat = targetFormat
	if conf.TargetFormat == 0 {
		conf.TargetFormat = i.ImageFormat
	}

	return i.doWithImageConfig(conf, func(src image.Image) (image.Image, error) {
		var filters []gift.Filter
		for _, f := range gfilters {
			f = valueobject.UnwrapFilter(f)
			if specProvider, ok := f.(valueobject.ImageProcessSpecProvider); ok {
				processSpec := specProvider.ImageProcessSpec()
				action, options := i.resolveActionOptions(processSpec)
				conf, err := valueobject.DecodeImageConfig(action, options, i.ImageService, i.ImageFormat)
				if err != nil {
					return nil, err
				}
				pFilters, err := i.Proc.FiltersFromConfig(src, conf)
				if err != nil {
					return nil, err
				}
				filters = append(filters, pFilters...)
			} else if orientationProvider, ok := f.(valueobject.ImageFilterFromOrientationProvider); ok {
				tf := orientationProvider.AutoOrient(i.Exif())
				if tf != nil {
					filters = append(filters, tf)
				}
			} else {
				filters = append(filters, f)
			}
		}
		return i.Proc.Filter(src, filters...)
	})
}

func (i *imageResource) resolveActionOptions(spec string) (string, []string) {
	var action string
	options := strings.Fields(spec)
	for i, p := range options {
		if hstrings.InSlicEqualFold(imageActions, p) {
			action = p
			options = append(options[:i], options[i+1:]...)
			break
		}
	}
	return action, options
}

func (i *imageResource) processActionSpec(action, spec string) (resources.ImageResource, error) {
	return i.processActionOptions(action, strings.Fields(spec))
}

func (i *imageResource) processActionOptions(action string, options []string) (resources.ImageResource, error) {
	conf, err := valueobject.DecodeImageConfig(action, options, i.ImageService, i.ImageFormat)
	if err != nil {
		return nil, err
	}

	img, err := i.doWithImageConfig(conf, func(src image.Image) (image.Image, error) {
		return i.Proc.ApplyFiltersFromConfig(src, conf)
	})
	if err != nil {
		return nil, err
	}

	if action == valueobject.ActionFill {
		if conf.Anchor == 0 && img.Width() == 0 || img.Height() == 0 {
			// See https://github.com/gohugoio/hugo/issues/7955
			// Smartcrop fails silently in some rare cases.
			// Fall back to a center fill.
			conf.Anchor = gift.CenterAnchor
			conf.AnchorStr = "center"
			return i.doWithImageConfig(conf, func(src image.Image) (image.Image, error) {
				return i.Proc.ApplyFiltersFromConfig(src, conf)
			})
		}
	}

	return img, nil
}

// Serialize images processing. The imaging library spins up its own set of Go routines,
// so there is not much to gain from adding more load to the mix. That
// can even have negative effect in low resource scenarios.
// Note that this only effects the non-cached scenario. Once the processed
// images is written to disk, everything is fast, fast fast.
const imageProcWorkers = 1

var imageProcSem = make(chan bool, imageProcWorkers)

func (i *imageResource) doWithImageConfig(conf valueobject.ImageConfig, f func(src image.Image) (image.Image, error)) (resources.ImageResource, error) {
	img, err := i.ImageCache.getOrCreate(i, conf, func() (*imageResource, image.Image, error) {
		imageProcSem <- true
		defer func() {
			<-imageProcSem
		}()

		src, err := i.DecodeImage()
		if err != nil {
			return nil, nil, &os.PathError{Op: conf.Action, Path: i.TargetPath(), Err: err}
		}

		converted, err := f(src)
		if err != nil {
			return nil, nil, &os.PathError{Op: conf.Action, Path: i.TargetPath(), Err: err}
		}

		hasAlpha := !valueobject.IsOpaque(converted)
		shouldFill := conf.BgColor != nil && hasAlpha
		shouldFill = shouldFill || (!valueobject.SupportsTransparency(conf.TargetFormat) && hasAlpha)
		var bgColor color.Color

		if shouldFill {
			bgColor = conf.BgColor
			if bgColor == nil {
				bgColor = i.ImageService.BgColor()
			}
			tmp := image.NewRGBA(converted.Bounds())
			draw.Draw(tmp, tmp.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)
			draw.Draw(tmp, tmp.Bounds(), converted, converted.Bounds().Min, draw.Over)
			converted = tmp
		}

		if conf.TargetFormat == resources.PNG {
			// Apply the colour palette from the source
			if paletted, ok := src.(*image.Paletted); ok {
				palette := paletted.Palette
				if bgColor != nil && len(palette) < 256 {
					palette = images.AddColorToPalette(bgColor, palette)
				} else if bgColor != nil {
					images.ReplaceColorInPalette(bgColor, palette)
				}
				tmp := image.NewPaletted(converted.Bounds(), palette)
				draw.FloydSteinberg.Draw(tmp, tmp.Bounds(), converted, converted.Bounds().Min)
				converted = tmp
			}
		}

		ci := i.clone(converted)
		targetPath := i.relTargetPathFromConfig(conf)
		ci.SetTargetPath(targetPath)
		ci.ImageFormat = conf.TargetFormat
		ci.SetMediaType(valueobject.MediaType(conf.TargetFormat))

		return ci, converted, nil
	})
	if err != nil {
		return nil, err
	}
	return img, nil
}

type giphy struct {
	image.Image
	gif *gif.GIF
}

func (g *giphy) GIF() *gif.GIF {
	return g.gif
}

// DecodeImage decodes the images source into an Image.
// This for internal use only.
func (i *imageResource) DecodeImage() (image.Image, error) {
	f, err := i.ReadSeekCloser()
	if err != nil {
		return nil, fmt.Errorf("failed to open images for decode: %w", err)
	}
	defer f.Close()

	if i.ImageFormat == resources.GIF {
		g, err := gif.DecodeAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to decode gif: %w", err)
		}
		return &giphy{gif: g, Image: g.Image[0]}, nil
	}
	img, _, err := image.Decode(f)
	return img, err
}

func (i *imageResource) clone(img image.Image) *imageResource {
	spec := i.baseResource.Clone().(baseResource)

	var image *valueobject.Image
	if img != nil {
		image = i.WithImage(img)
	} else {
		image = i.WithSpec(spec)
	}

	return &imageResource{
		Image:        image,
		root:         i.root,
		baseResource: spec,
	}
}

func (i *imageResource) getImageMetaCacheTargetPath() string {
	const imageMetaVersionNumber = 1 // Increment to invalidate the meta cache

	cfgHash := i.ImageService.SourceHash()
	df := i.baseResource.getResourcePaths() //TODO
	p1, _ := paths.FileAndExt(df.File)
	h := i.Hash()
	idStr := identity.HashString(h, i.Size(), imageMetaVersionNumber, cfgHash)
	df.File = fmt.Sprintf("%s_%s.json", p1, idStr)
	return df.TargetPath()
}

func (i *imageResource) relTargetPathFromConfig(conf valueobject.ImageConfig) valueobject.ResourcePaths {
	p1, p2 := paths.FileAndExt(i.getResourcePaths().File)
	if conf.TargetFormat != i.ImageFormat {
		p2 = valueobject.DefaultExtension(conf.TargetFormat)
	}

	h := i.Hash()
	idStr := fmt.Sprintf("_hu%s_%d", h, i.Size())

	// Do not change for no good reason.
	const md5Threshold = 100

	key := conf.GetKey(i.ImageFormat)

	// It is useful to have the key in clear text, but when nesting transforms, it
	// can easily be too long to read, and maybe even too long
	// for the different OSes to handle.
	if len(p1)+len(idStr)+len(p2) > md5Threshold {
		key = helpers.MD5String(p1 + key + p2)
		huIdx := strings.Index(p1, "_hu")
		if huIdx != -1 {
			p1 = p1[:huIdx]
		} else {
			// This started out as a very long file name. Making it even longer
			// could melt ice in the Arctic.
			p1 = ""
		}
	} else if strings.Contains(p1, idStr) {
		// On scaling an already scaled images, we get the file info from the original.
		// Repeating the same info in the filename makes it stuttery for no good reason.
		idStr = ""
	}

	rp := i.getResourcePaths()
	rp.File = fmt.Sprintf("%s%s_%s%s", p1, idStr, key, p2)

	return rp
}
