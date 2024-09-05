package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/entity"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func New(services contenthub.Services) (*entity.ContentHub, error) {
	log := loggers.NewDefault()

	cs, err := newContentSpec()
	if err != nil {
		return nil, err
	}

	cache := newCache()

	ch := &entity.ContentHub{
		Cache:            cache,
		Fs:               services,
		TemplateExecutor: nil,
		PageMap: &entity.PageMap{
			PageTrees: newPageTree(),

			PageBuilder: &entity.PageBuilder{
				LangSvc:     services,
				TaxonomySvc: services,
				TemplateSvc: nil, // TODO, set when used

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
		Title: &entity.Title{
			Style: entity.StyleAP,
		},
		Log: log,
	}

	return ch, nil
}

func newCache() *entity.Cache {
	memCache := dynacache.New(dynacache.Options{Running: true, Log: loggers.NewDefault()})

	return &entity.Cache{
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
	treeConfig := doctree.Config[*entity.PageTreesNode]{
		Shifter: &entity.SourceShifter{},
	}

	pageTrees := &entity.PageTrees{
		TreePages: doctree.New(
			treeConfig,
		),
		TreeResources: doctree.New(
			treeConfig,
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
