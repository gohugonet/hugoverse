package factory

import (
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/internal/domain/resources/entity"
	"github.com/mdfriday/hugoverse/pkg/media"
	"github.com/mdfriday/hugoverse/pkg/output"
	"github.com/tdewolff/minify/v2"
	"regexp"
)

// NewMinifier creates a new Client given a specification. Note that it is the media types
// configured for the site that is used to match files to the correct minifier.
func NewMinifier(mts media.Types, ofs output.Formats, minifyService resources.MinifyConfig) (*entity.MinifierClient, error) {
	m := minify.New()

	minifyService.Minifiers(mts, func(mt media.Type, min minify.Minifier) {
		m.Add(mt.Type, min)
	})

	m.AddRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), minifyService.GetMinifier("js"))
	m.AddRegexp(regexp.MustCompile(`^(application|text)/(x-|(ld|manifest)\+)?json$`), minifyService.GetMinifier("json"))

	for _, of := range ofs {
		if of.IsHTML {
			m.Add(of.MediaType.Type, minifyService.GetMinifier("html"))
		}
	}

	return &entity.MinifierClient{M: m, MinifyOutput: minifyService.IsMinifyPublish()}, nil
}
