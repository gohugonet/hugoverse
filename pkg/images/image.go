package images

import (
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"strings"
)

var AnchorPositions = map[string]gift.Anchor{
	strings.ToLower("Center"):      gift.CenterAnchor,
	strings.ToLower("TopLeft"):     gift.TopLeftAnchor,
	strings.ToLower("Top"):         gift.TopAnchor,
	strings.ToLower("TopRight"):    gift.TopRightAnchor,
	strings.ToLower("Left"):        gift.LeftAnchor,
	strings.ToLower("Right"):       gift.RightAnchor,
	strings.ToLower("BottomLeft"):  gift.BottomLeftAnchor,
	strings.ToLower("Bottom"):      gift.BottomAnchor,
	strings.ToLower("BottomRight"): gift.BottomRightAnchor,
}

// These encoding hints are currently only relevant for Webp.
var Hints = map[string]webpoptions.EncodingPreset{
	"picture": webpoptions.EncodingPresetPicture,
	"photo":   webpoptions.EncodingPresetPhoto,
	"drawing": webpoptions.EncodingPresetDrawing,
	"icon":    webpoptions.EncodingPresetIcon,
	"text":    webpoptions.EncodingPresetText,
}
