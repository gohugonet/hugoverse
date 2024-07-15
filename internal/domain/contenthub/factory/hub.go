package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/entity"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func New(services contenthub.Services) (*entity.ContentHub, error) {
	log := loggers.NewDefault()

	cs, err := newContentSpec()
	if err != nil {
		return nil, err
	}

	cache := valueobject.NewCache()

	ch := &entity.ContentHub{
		Fs:               services,
		TemplateExecutor: nil,
		PageMap: &entity.PageMap{
			ContentSpec: cs,
			PageTrees:   newPageTree(),

			PageBuilder: &entity.PageBuilder{
				LangSvc: services,
				Taxonomy: &entity.Taxonomy{
					Views: services.Views(),
					FsSvc: services,
					Cache: cache,
				},
				TemplateSvc: nil, // TODO, set when used
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
		TreeTaxonomyEntries: doctree.NewTreeShiftTree[contenthub.WeightedContentNode](
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
