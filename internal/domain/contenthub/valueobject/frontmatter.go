package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/cast"
	"strings"
)

type FrontMatter struct {
	*Cascade

	Path string
	Lang string
	Kind string
}

type FrontMatterParser struct {
	Params      maps.Params
	LangService contenthub.LangService
}

func (b *FrontMatterParser) Parse() (*FrontMatter, error) {
	fm := &FrontMatter{}

	if err := b.parseCascade(fm); err != nil {
		return nil, err
	}

	if err := b.parseCustomized(fm); err != nil {
		return fm, err
	}

	return fm, nil
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
		if b.LangService.IsLanguageValid(lang) {
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
