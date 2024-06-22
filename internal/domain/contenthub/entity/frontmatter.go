package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/cast"
	"strings"
)

type FrontMatter struct {
	Params maps.Params
	*valueobject.Cascade

	// User defined params.
	Customized maps.Params
	Path       string
	Lang       string
	Kind       string

	langService contenthub.LangService
}

func (fm *FrontMatter) frontMatterMap(it pageparser.Item, source []byte) error {
	f := pageparser.FormatFromFrontMatterType(it.Type)

	m, err := metadecoders.Default.UnmarshalToMap(it.Val(source), f)

	if err != nil {
		return err
	}

	maps.PrepareParams(m)

	fm.Params = m

	return nil
}

func (fm *FrontMatter) parse() error {
	if err := fm.parseCascade(); err != nil {
		return err
	}
	return nil
}

func (fm *FrontMatter) parseCascade() error {
	if cv, found := fm.Params["cascade"]; found {
		var err error
		fm.Cascade, err = valueobject.NewCascade(cv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *FrontMatter) parseCustomized() error {
	// Look for path, lang and kind, all of which values we need early on.
	if v, found := fm.Params["path"]; found {
		fm.Path = paths.ToSlashPreserveLeading(cast.ToString(v))
		fm.Params["path"] = fm.Path
	}
	if v, found := fm.Params["lang"]; found {
		lang := strings.ToLower(cast.ToString(v))
		if fm.langService.IsLanguageValid(lang) {
			fm.Lang = lang
			fm.Params["lang"] = fm.Lang
		}
	}
	if v, found := fm.Params["kind"]; found {
		s := cast.ToString(v)
		if s != "" {
			fm.Kind = valueobject.GetKindMain(s)
			if fm.Kind == "" {
				return fmt.Errorf("unknown kind %q in front matter", s)
			}
			fm.Params["kind"] = fm.Kind
		}
	}
	return nil
}
