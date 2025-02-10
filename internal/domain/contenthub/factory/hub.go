package factory

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/entity"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/pkg/cache/dynacache"
	"github.com/mdfriday/hugoverse/pkg/cache/stale"
	"github.com/mdfriday/hugoverse/pkg/doctree"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/paths"
	"golang.org/x/text/language"
	"strings"

	"encoding/json"
	"github.com/gohugoio/go-i18n/v2/i18n"
	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v2"
)

func New(services contenthub.Services) (*entity.ContentHub, error) {
	log := loggers.NewDefault()
	valueobject.SetupDefaultContentTypes()

	cs, err := newContentSpec()
	if err != nil {
		return nil, err
	}

	t, err := newTranslator(services, log)
	if err != nil {
		return nil, err
	}

	cache := newCache()

	ch := &entity.ContentHub{
		Cache:            cache,
		Search:           entity.NewSearch(log),
		Fs:               services,
		TemplateExecutor: nil,
		Translator:       t,

		PageMap: &entity.PageMap{
			PageTrees: newPageTree(),

			PageBuilder: &entity.PageBuilder{
				LangSvc:     services,
				TaxonomySvc: services,
				MediaSvc:    services,
				TemplateSvc: nil, // TODO, set when used
				PageMapper:  nil,

				Taxonomy: &entity.Taxonomy{
					Views: services.Views(),
					FsSvc: services,
					Cache: cache,
				},
				Term: &entity.Term{
					Terms: nil,
					FsSvc: services,
					Cache: cache,
				},
				Section:    &entity.Section{FsSvc: services, Cache: cache},
				Standalone: &entity.Standalone{FsSvc: services, Cache: cache},

				ConvertProvider: cs,

				Log: log,
			},

			Cache: cache,
			Log:   log,
		},
		PageFinder: &entity.PageFinder{
			Fs: services,
		},

		Title: &entity.Title{
			Style: entity.StyleAP,
		},
		Log: log,
	}

	ch.PageMap.SetupReverseIndex()
	ch.PageBuilder.PageMapper = ch.PageMap
	ch.PageBuilder.ContentHub = ch
	ch.PageFinder.PageMapper = ch.PageMap

	return ch, nil
}

func newCache() *entity.Cache {
	memCache := dynacache.New(dynacache.Options{Running: true, Log: loggers.NewDefault()})

	return &entity.Cache{
		CachePages1: dynacache.GetOrCreatePartition[string, contenthub.Pages](
			memCache, "/pag1",
			dynacache.OptionsPartition{Weight: 10, ClearWhen: dynacache.ClearOnRebuild},
		),
		CachePages2: dynacache.GetOrCreatePartition[string, contenthub.Pages](
			memCache, "/pag2",
			dynacache.OptionsPartition{Weight: 10, ClearWhen: dynacache.ClearOnRebuild},
		),
		CacheContentSource: dynacache.GetOrCreatePartition[string, *stale.Value[[]byte]](
			memCache, "/cont/src",
			dynacache.OptionsPartition{Weight: 70, ClearWhen: dynacache.ClearOnChange},
		),
		CachePageSource: dynacache.GetOrCreatePartition[string, contenthub.PageSource](
			memCache,
			"/page/source",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 40},
		),
		CachePageSources: dynacache.GetOrCreatePartition[string, []contenthub.PageSource](
			memCache,
			"/page/sources",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnRebuild, Weight: 40},
		),

		CacheContentRendered: dynacache.GetOrCreatePartition[string, *stale.Value[valueobject.ContentSummary]](
			memCache,
			"/cont/ren",
			dynacache.OptionsPartition{Weight: 70, ClearWhen: dynacache.ClearOnChange},
		),

		CacheContentToRender: dynacache.GetOrCreatePartition[string, *stale.Value[[]byte]](
			memCache,
			"/cont/toc",
			dynacache.OptionsPartition{Weight: 70, ClearWhen: dynacache.ClearOnChange},
		),

		CacheContentShortcodes: dynacache.GetOrCreatePartition[string, *stale.Value[map[string]valueobject.ShortcodeRenderer]](
			memCache,
			"/cont/shortcodes",
			dynacache.OptionsPartition{Weight: 70, ClearWhen: dynacache.ClearOnChange},
		),
	}
}

// newContentSpec returns a ContentSpec initialized
// with the appropriate fields from the given config.Provider.
func newContentSpec() (*entity.ContentSpec, error) {
	spec := &entity.ContentSpec{}

	// markdown
	converterRegistry, err := newConverterRegistry()
	if err != nil {
		return nil, err
	}

	spec.Converters = converterRegistry

	return spec, nil
}

func newPageTree() *entity.PageTrees {
	pageTrees := &entity.PageTrees{
		TreePages: doctree.New(
			doctree.Config[*entity.PageTreesNode]{
				Shifter: &entity.PageShifter{Shifter: &entity.Shifter{}},
			},
		),
		TreeResources: doctree.New(
			doctree.Config[*entity.PageTreesNode]{
				Shifter: &entity.SourceShifter{Shifter: &entity.Shifter{}},
			},
		),
		TreeTaxonomyEntries: doctree.NewTreeShiftTree[*entity.WeightedTermTreeNode](
			doctree.DimensionLanguage.Index(), 2), // TODO: get this from config
	}

	pageTrees.CreateMutableTrees()

	return pageTrees
}

func newConverterRegistry() (contenthub.ConverterRegistry, error) {
	converters := make(map[string]contenthub.ConverterProvider)

	add := func(p contenthub.ProviderProvider) error {
		c, err := p.New()
		if err != nil {
			return err
		}

		name := c.Name()
		converters[name] = c

		return nil
	}

	// default
	if err := add(valueobject.MDProvider); err != nil {
		return nil, err
	}

	return &valueobject.ConverterRegistry{
		Converters: converters,
	}, nil
}

func newTranslator(services contenthub.Services, log loggers.Logger) (*entity.Translator, error) {
	defaultLangTag, err := language.Parse(services.DefaultLanguage())
	if err != nil {
		defaultLangTag = language.English
	}
	bundle := i18n.NewBundle(defaultLangTag)

	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	if err := services.WalkI18n("", fs.WalkCallback{
		HookPre: nil,
		WalkFn: func(path string, info fs.FileMetaInfo) error {
			if info.IsDir() {
				return nil
			}
			file, err := valueobject.NewFileInfo(info)
			if err != nil {
				return err
			}
			return addTranslationFile(bundle, file)
		},
		HookPost: nil,
	}, fs.WalkwayConfig{}); err != nil {
		if !herrors.IsNotExist(err) {
			return nil, err
		}
	}

	t := &entity.Translator{
		ContentLanguage: services.DefaultLanguage(),
		TranslateFuncs:  make(map[string]entity.TranslateFunc),

		Log: log,
	}
	t.SetupTranslateFuncs(bundle)

	return t, err
}

func addTranslationFile(bundle *i18n.Bundle, r *valueobject.File) error {
	f, err := r.Open()
	if err != nil {
		return fmt.Errorf("failed to open translations file %q:: %w", r.LogicalName(), err)
	}

	b := helpers.ReaderToBytes(f)
	f.Close()

	name := r.LogicalName()
	lang := paths.Filename(name)
	tag := language.Make(lang)
	if tag == language.Und {
		try := entity.ArtificialLangTagPrefix + lang
		_, err = language.Parse(try)
		if err != nil {
			return fmt.Errorf("%q: %s", try, err)
		}
		name = entity.ArtificialLangTagPrefix + name
	}

	_, err = bundle.ParseMessageFileBytes(b, name)
	if err != nil {
		if strings.Contains(err.Error(), "no plural rule") {
			// https://github.com/gohugoio/hugo/issues/7798
			name = entity.ArtificialLangTagPrefix + name
			_, err = bundle.ParseMessageFileBytes(b, name)
			if err == nil {
				return nil
			}
		}
		return errWithFileContext(fmt.Errorf("failed to load translations: %w", err), r)
	}

	return nil
}

func errWithFileContext(inerr error, r *valueobject.File) error {
	realFilename := r.Filename()
	f, err := r.Open()
	if err != nil {
		return inerr
	}
	defer f.Close()

	return herrors.NewFileErrorFromName(inerr, realFilename).UpdateContent(f, nil)
}
