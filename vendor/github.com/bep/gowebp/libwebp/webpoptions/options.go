package webpoptions

const (
	EncodingPresetDefault EncodingPreset = iota
	EncodingPresetPicture
	EncodingPresetPhoto
	EncodingPresetDrawing
	EncodingPresetIcon
	EncodingPresetText
)

type (
	EncodingPreset  int
	EncodingOptions struct {
		// Quality is a number between 0 and 100. Set to 0 for lossless.
		Quality int

		// The encoding preset to use.
		EncodingPreset

		// Use sharp (and slow) RGB->YUV conversion.
		UseSharpYuv bool
	}
)
