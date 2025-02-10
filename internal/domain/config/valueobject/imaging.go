package valueobject

import (
	"errors"
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/config"
	"github.com/mdfriday/hugoverse/pkg/identity"
	"github.com/mdfriday/hugoverse/pkg/images"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"strings"
)

const (
	defaultJPEGQuality    = 75
	defaultResampleFilter = "box"
	defaultBgColor        = "#ffffff"
	defaultHint           = "photo"
)

var (
	defaultImagingConfig = map[string]any{
		"resampleFilter": defaultResampleFilter,
		"bgColor":        defaultBgColor,
		"hint":           defaultHint,
		"quality":        defaultJPEGQuality,
	}
)

const (
	// Do not change.
	smartCropIdentifier = "smart"

	// This is just a increment, starting on 1. If Smart Crop improves its cropping, we
	// need a way to trigger a re-generation of the crops in the wild, so increment this.
	smartCropVersionNumber = 1
)

// ImagingConfig contains default images processing configuration. This will be fetched
// from site (or language) config.
type ImagingConfig struct {
	// Default images quality setting (1-100). Only used for JPEG images.
	Quality int

	// Resample filter to use in resize operations.
	ResampleFilter string

	// Hint about what type of images this is.
	// Currently only used when encoding to Webp.
	// Default is "photo".
	// Valid values are "picture", "photo", "drawing", "icon", or "text".
	Hint string

	// The anchor to use in Fill. Default is "smart", i.e. Smart Crop.
	Anchor string

	// Default color used in fill operations (e.g. "fff" for white).
	BgColor string

	Exif ExifConfig
}

func DecodeImagingConfig(p config.Provider) (ImagingConfigInternal, error) {
	in := make(map[string]any)

	if p.IsSet("imaging") {
		in = p.GetStringMap("imaging")
	}

	buildConfig := func(in any) (ImagingConfigInternal, error) {
		m, err := maps.ToStringMapE(in)
		if err != nil {
			return ImagingConfigInternal{}, err
		}
		// Merge in the defaults.
		maps.MergeShallow(m, defaultImagingConfig)

		var i ImagingConfigInternal
		if err := mapstructure.Decode(m, &i.Imaging); err != nil {
			return i, err
		}

		if err := i.Imaging.init(); err != nil {
			return i, err
		}

		i.BgColor, err = images.HexStringToColor(i.Imaging.BgColor)
		if err != nil {
			return i, err
		}

		if i.Imaging.Anchor != "" && i.Imaging.Anchor != smartCropIdentifier {
			anchor, found := anchorPositions[i.Imaging.Anchor]
			if !found {
				return i, fmt.Errorf("invalid anchor value %q in imaging config", i.Anchor)
			}
			i.Anchor = anchor
		}

		filter, found := images.ImageFilters[i.Imaging.ResampleFilter]
		if !found {
			return i, fmt.Errorf("%q is not a valid resample filter", filter)
		}

		i.ResampleFilter = filter

		return i, nil
	}

	// Build the config
	c, err := buildConfig(in)
	if err != nil {
		return ImagingConfigInternal{}, err
	}

	c.SourceHash = identity.HashString(in)
	return c, nil
}

func (cfg *ImagingConfig) init() error {
	if cfg.Quality < 0 || cfg.Quality > 100 {
		return errors.New("images quality must be a number between 1 and 100")
	}

	cfg.BgColor = strings.ToLower(strings.TrimPrefix(cfg.BgColor, "#"))
	cfg.Anchor = strings.ToLower(cfg.Anchor)
	cfg.ResampleFilter = strings.ToLower(cfg.ResampleFilter)
	cfg.Hint = strings.ToLower(cfg.Hint)

	if cfg.Anchor == "" {
		cfg.Anchor = smartCropIdentifier
	}

	if strings.TrimSpace(cfg.Exif.IncludeFields) == "" && strings.TrimSpace(cfg.Exif.ExcludeFields) == "" {
		// Don't change this for no good reason. Please don't.
		cfg.Exif.ExcludeFields = "GPS|Exif|Exposure[M|P|B]|Contrast|Resolution|Sharp|JPEG|Metering|Sensing|Saturation|ColorSpace|Flash|WhiteBalance"
	}

	return nil
}
