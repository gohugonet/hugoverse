package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/output"
	"strings"
)

// OutputFormats holds a list of the relevant output formats for a given page.
type OutputFormats []OutputFormat

// Get gets a OutputFormat given its name, i.e. json, html etc.
// It returns nil if none found.
func (o OutputFormats) Get(name string) *OutputFormat {
	for _, f := range o {
		if strings.EqualFold(f.Format.Name, name) {
			return &f
		}
	}
	return nil
}

// OutputFormat links to a representation of a resource.
type OutputFormat struct {
	// Rel contains a value that can be used to construct a rel link.
	// This is value is fetched from the output format definition.
	// Note that for pages with only one output format,
	// this method will always return "canonical".
	// As an example, the AMP output format will, by default, return "amphtml".
	//
	// See:
	// https://www.ampproject.org/docs/guides/deploy/discovery
	//
	// Most other output formats will have "alternate" as value for this.
	Rel string

	Format output.Format

	relPermalink string
	permalink    string
}

func NewOutputFormat(relPermalink, permalink string, f output.Format) OutputFormat {
	return OutputFormat{Rel: "canonical", Format: f, relPermalink: relPermalink, permalink: permalink}
}

// Permalink returns the absolute permalink to this output format.
func (o OutputFormat) Permalink() string {
	return o.permalink
}

// RelPermalink returns the relative permalink to this output format.
func (o OutputFormat) RelPermalink() string {
	return o.relPermalink
}
