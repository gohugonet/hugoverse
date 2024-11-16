package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/types"
	"github.com/spf13/cast"
	"strings"
	"time"
)

type FrontMatter struct {
	*Cascade

	Path   string
	Lang   string
	Kind   string
	Title  string
	Weight int

	Terms map[string][]string

	Params maps.Params
}

type FrontMatterParser struct {
	Params      maps.Params
	LangSvc     contenthub.LangService
	TaxonomySvc contenthub.TaxonomyService
}

func (b *FrontMatterParser) Parse() (*FrontMatter, error) {
	fm := &FrontMatter{
		Terms:  map[string][]string{},
		Params: b.Params,
	}

	if err := b.parseCascade(fm); err != nil {
		return nil, err
	}

	if err := b.parseCustomized(fm); err != nil {
		return fm, err
	}

	if err := b.parseTerms(fm); err != nil {
		return nil, err
	}

	if err := b.parseTitle(fm); err != nil {
		return nil, err
	}

	if err := b.parseWeight(fm); err != nil {
		return nil, err
	}

	return fm, nil
}

func (b *FrontMatterParser) parseWeight(fm *FrontMatter) error {
	fm.Weight = 0
	if v, found := b.Params["weight"]; found {
		fm.Weight = cast.ToInt(v)
	}
	return nil
}

func (b *FrontMatterParser) parseTitle(fm *FrontMatter) error {
	if v, found := b.Params["title"]; found {
		fm.Title = cast.ToString(v)
	}
	return nil
}

func (b *FrontMatterParser) parseCascade(fm *FrontMatter) error {
	if cv, found := b.Params["cascade"]; found {
		var err error
		fm.Cascade, err = NewCascade(cv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *FrontMatterParser) parseCustomized(fm *FrontMatter) error {
	// Look for path, lang and kind, all of which values we need early on.
	if v, found := b.Params["path"]; found {
		fm.Path = paths.ToSlashPreserveLeading(cast.ToString(v))
	}
	if v, found := b.Params["lang"]; found {
		lang := strings.ToLower(cast.ToString(v))
		if b.LangSvc.IsLanguageValid(lang) {
			fm.Lang = lang
		}
	}
	if v, found := b.Params["kind"]; found {
		s := cast.ToString(v)
		if s != "" {
			fm.Kind = GetKindMain(s)
			if fm.Kind == "" {
				return fmt.Errorf("unknown kind %q in front matter", s)
			}
		}
	}
	return nil
}

func (b *FrontMatterParser) parseTerms(fm *FrontMatter) error {
	views := b.TaxonomySvc.Views()

	for _, viewName := range views {
		vals := types.ToStringSlicePreserveString(GetParam(b.Params, viewName.Plural(), false))
		if vals == nil {
			continue
		}

		fm.Terms[viewName.Plural()] = vals
	}
	return nil
}

func GetParamToLower(m maps.Params, key string) any {
	return GetParam(m, key, true)
}

func GetParam(p maps.Params, key string, stringToLower bool) any {
	v := p[strings.ToLower(key)]

	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case bool:
		return val
	case string:
		if stringToLower {
			return strings.ToLower(val)
		}
		return val
	case int64, int32, int16, int8, int:
		return cast.ToInt(v)
	case float64, float32:
		return cast.ToFloat64(v)
	case time.Time:
		return val
	case []string:
		if stringToLower {
			return helpers.SliceToLower(val)
		}
		return v
	default:
		return v
	}
}
