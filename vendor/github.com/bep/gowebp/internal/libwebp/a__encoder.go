package libwebp

/*
#include <stdlib.h>
#include <string.h> // for memset
#ifndef LIBWEBP_NO_SRC
#include <encode.h>
#else
#include <webp/encode.h>
#endif

static uint8_t* encodeNRGBA(WebPConfig* config, const uint8_t* rgba, int width, int height, int stride, size_t* output_size) {
	WebPPicture pic;
	WebPMemoryWriter wrt;
	int ok;
	if (!WebPPictureInit(&pic)) {
		return NULL;
	}
	pic.use_argb = 1;
	pic.width = width;
	pic.height = height;
	pic.writer = WebPMemoryWrite;
	pic.custom_ptr = &wrt;
	WebPMemoryWriterInit(&wrt);
	ok = WebPPictureImportRGBA(&pic, rgba, stride) && WebPEncode(config, &pic);
	WebPPictureFree(&pic);
	if (!ok) {
		WebPMemoryWriterClear(&wrt);
		return NULL;
	}
	*output_size = wrt.size;
	return wrt.mem;
}

static uint8_t* encodeGray(WebPConfig* config, uint8_t *y, int width, int height, int stride, size_t* output_size) {
	WebPPicture pic;
	WebPMemoryWriter wrt;

	int ok;
	if (!WebPPictureInit(&pic)) {
		return NULL;
	}

	pic.use_argb = 0;
	pic.width = width;
	pic.height = height;
	pic.y_stride = stride;
	pic.writer = WebPMemoryWrite;
	pic.custom_ptr = &wrt;
	WebPMemoryWriterInit(&wrt);

	const int uvWidth = (int)(((int64_t)width + 1) >> 1);
  	const int uvHeight = (int)(((int64_t)height + 1) >> 1);
  	const int uvStride = uvWidth;
	const int uvSize = uvStride * uvHeight;
	const int gray = 128;
	uint8_t* chroma;

	chroma = malloc(uvSize);
	if (!chroma) {
		return 0;
	}
	memset(chroma, gray, uvSize);

	pic.y = y;
	pic.u = chroma;
	pic.v = chroma;
	pic.uv_stride = uvStride;

	ok = WebPEncode(config, &pic);

	free(chroma);

	WebPPictureFree(&pic);
	if (!ok) {
		WebPMemoryWriterClear(&wrt);
		return NULL;
	}
	*output_size = wrt.size;
	return wrt.mem;

}

*/
import "C"

import (
	"errors"
	"image"
	"image/draw"
	"io"
	"unsafe"

	"github.com/bep/gowebp/libwebp/webpoptions"
)

type (
	Encoder struct {
		config *C.WebPConfig
		img    *image.NRGBA
	}
)

// Encode encodes src into w considering the options in o.
//
// Any src that isn't one of *image.RGBA, *image.NRGBA, or *image.Gray
// will be converted to *image.NRGBA using draw.Draw first.
//
func Encode(w io.Writer, src image.Image, o webpoptions.EncodingOptions) error {
	config, err := encodingOptionsToCConfig(o)
	if err != nil {
		return err
	}

	var (
		bounds = src.Bounds()
		output *C.uchar
		size   C.size_t
	)

	switch v := src.(type) {
	case *image.RGBA:
		output = C.encodeNRGBA(
			config,
			(*C.uint8_t)(&v.Pix[0]),
			C.int(bounds.Max.X),
			C.int(bounds.Max.Y),
			C.int(v.Stride),
			&size,
		)
	case *image.NRGBA:
		output = C.encodeNRGBA(
			config,
			(*C.uint8_t)(&v.Pix[0]),
			C.int(bounds.Max.X),
			C.int(bounds.Max.Y),
			C.int(v.Stride),
			&size,
		)
	case *image.Gray:
		gray := (*C.uint8_t)(&v.Pix[0])
		output = C.encodeGray(
			config,
			gray,
			C.int(bounds.Max.X),
			C.int(bounds.Max.Y),
			C.int(v.Stride),
			&size,
		)
	default:
		rgba := ConvertToNRGBA(src)
		output = C.encodeNRGBA(
			config,
			(*C.uint8_t)(&rgba.Pix[0]),
			C.int(bounds.Max.X),
			C.int(bounds.Max.Y),
			C.int(rgba.Stride),
			&size,
		)
	}

	if output == nil || size == 0 {
		return errors.New("failed to encode")
	}
	defer C.free(unsafe.Pointer(output))

	_, err = w.Write(((*[1 << 30]byte)(unsafe.Pointer(output)))[0:int(size):int(size)])

	return err
}

func ConvertToNRGBA(src image.Image) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)

	return dst
}

func encodingOptionsToCConfig(o webpoptions.EncodingOptions) (*C.WebPConfig, error) {
	cfg := &C.WebPConfig{}
	quality := C.float(o.Quality)

	if C.WebPConfigPreset(cfg, C.WebPPreset(o.EncodingPreset), quality) == 0 {
		return nil, errors.New("failed to init encoder config")
	}

	if quality == 0 {
		// Activate the lossless compression mode with the desired efficiency level
		// between 0 (fastest, lowest compression) and 9 (slower, best compression).
		// A good default level is '6', providing a fair tradeoff between compression
		// speed and final compressed size.
		if C.WebPConfigLosslessPreset(cfg, C.int(6)) == 0 {
			return nil, errors.New("failed to init lossless preset")
		}
	}

	cfg.use_sharp_yuv = boolToCInt(o.UseSharpYuv)

	if C.WebPValidateConfig(cfg) == 0 {
		return nil, errors.New("failed to validate config")
	}

	return cfg, nil
}

func boolToCInt(b bool) (result C.int) {
	result = 0

	if b {
		result = 1
	}

	return
}
