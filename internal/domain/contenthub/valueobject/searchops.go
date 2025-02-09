package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/types"
	"github.com/spf13/cast"
	"strings"
	"time"
)

const (
	TypeBasic     = "basic"
	TypeFragments = "fragments"
)

// SearchOpts holds the options for a related search.
type SearchOpts struct {
	// The Document to search for related content for.
	Document contenthub.Document

	// The keywords to search for.
	NamedSlices []types.KeyValues

	// The indices to search in.
	Indices []string

	// Fragments holds a a list of special keywords that is used
	// for indices configured as type "fragments".
	// This will match the fragment identifiers of the documents.
	Fragments []string
}

// StringKeyword is a string search keyword.
type StringKeyword string

func (s StringKeyword) String() string {
	return string(s)
}

// FragmentKeyword represents a document fragment.
type FragmentKeyword string

func (f FragmentKeyword) String() string {
	return string(f)
}

// IndexConfig configures an index.
type IndexConfig struct {
	// The index name. This directly maps to a field or Param name.
	IndexName string

	// The index type.
	IndexType string

	// Enable to apply a type specific filter to the results.
	// This is currently only used for the "fragments" type.
	IndexApplyFilter bool

	// Contextual pattern used to convert the Param value into a string.
	// Currently only used for dates. Can be used to, say, bump posts in the same
	// time frame when searching for related documents.
	// For dates it follows Go's time.Format patterns, i.e.
	// "2006" for YYYY and "200601" for YYYYMM.
	IndexPattern string

	// This field's weight when doing multi-index searches. Higher is "better".
	IndexWeight int

	// A percentage (0-100) used to remove common keywords from the index.
	// As an example, setting this to 50 will remove all keywords that are
	// used in more than 50% of the documents in the index.
	IndexCardinalityThreshold int

	// Will lower case all string values in and queries tothis index.
	// May get better accurate results, but at a slight performance cost.
	IndexToLower bool
}

func (cfg IndexConfig) Name() string {
	return cfg.IndexName
}
func (cfg IndexConfig) Type() string {
	return cfg.IndexType
}
func (cfg IndexConfig) ApplyFilter() bool {
	return cfg.IndexApplyFilter
}
func (cfg IndexConfig) Pattern() string {
	return cfg.IndexPattern
}
func (cfg IndexConfig) Weight() int {
	return cfg.IndexWeight
}
func (cfg IndexConfig) CardinalityThreshold() int {
	return cfg.IndexCardinalityThreshold
}
func (cfg IndexConfig) ToLower() bool {
	return cfg.IndexToLower
}

func (cfg IndexConfig) stringToKeyword(s string) contenthub.Keyword {
	if cfg.ToLower() {
		s = strings.ToLower(s)
	}
	if cfg.Type() == TypeFragments {
		return FragmentKeyword(s)
	}
	return StringKeyword(s)
}

// ToKeywords returns a Keyword slice of the given input.
func (cfg IndexConfig) ToKeywords(v any) ([]contenthub.Keyword, error) {
	var keywords []contenthub.Keyword

	switch vv := v.(type) {
	case string:
		keywords = append(keywords, cfg.stringToKeyword(vv))
	case []string:
		vvv := make([]contenthub.Keyword, len(vv))
		for i := 0; i < len(vvv); i++ {
			vvv[i] = cfg.stringToKeyword(vv[i])
		}
		keywords = append(keywords, vvv...)
	case []any:
		return cfg.ToKeywords(cast.ToStringSlice(vv))
	case time.Time:
		layout := "2006"
		if cfg.Pattern() != "" {
			layout = cfg.Pattern()
		}
		keywords = append(keywords, StringKeyword(vv.Format(layout)))
	case nil:
		return keywords, nil
	default:
		return keywords, fmt.Errorf("indexing currently not supported for index %q and type %T", cfg.Name(), vv)
	}

	return keywords, nil
}
