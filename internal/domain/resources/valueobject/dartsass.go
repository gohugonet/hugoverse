package valueobject

import (
	"github.com/mdfriday/hugoverse/pkg/hexec"
	"github.com/mdfriday/hugoverse/pkg/paths"
	"github.com/mitchellh/mapstructure"
	"os"
	"strings"
)

var DartSassBinaryName string

func SetDartSassBinaryName() {
	DartSassBinaryName = os.Getenv("DART_SASS_BINARY")
	if DartSassBinaryName == "" {
		for _, name := range dartSassBinaryNamesV2 {
			if hexec.InPath(name) {
				DartSassBinaryName = name
				break
			}
		}
		if DartSassBinaryName == "" {
			if hexec.InPath(dartSassBinaryNameV1) {
				DartSassBinaryName = dartSassBinaryNameV1
			}
		}
	}
}

var (
	dartSassBinaryNameV1  = "dart-sass-embedded"
	dartSassBinaryNamesV2 = []string{"dart-sass", "sass"}
)

func IsDartSassV2() bool {
	return !strings.Contains(DartSassBinaryName, "embedded")
}

type DartSassOptions struct {
	// Hugo, will by default, just replace the extension of the source
	// to .css, e.g. "scss/main.scss" becomes "scss/main.css". You can
	// control this by setting this, e.g. "styles/main.css" will create
	// a Resource with that as a base for RelPermalink etc.
	TargetPath string

	// Hugo automatically adds the entry directories (where the main.scss lives)
	// for project and themes to the list of include paths sent to LibSASS.
	// Any paths set in this setting will be appended. Note that these will be
	// treated as relative to the working dir, i.e. no include paths outside the
	// project/themes.
	IncludePaths []string

	// Default is nested.
	// One of nested, expanded, compact, compressed.
	OutputStyle string

	// When enabled, Hugo will generate a source map.
	EnableSourceMap bool

	// If enabled, sources will be embedded in the generated source map.
	SourceMapIncludeSources bool

	// Vars will be available in 'hugo:vars', e.g:
	//     @use "hugo:vars";
	//     $color: vars.$color;
	Vars map[string]any
}

func DecodeDartSassOptions(m map[string]any) (opts DartSassOptions, err error) {
	if m == nil {
		return
	}
	err = mapstructure.WeakDecode(m, &opts)

	if opts.TargetPath != "" {
		opts.TargetPath = paths.ToSlashTrimLeading(opts.TargetPath)
	}

	return
}
