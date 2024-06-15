package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/entity"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/radixtree"
)

func New(fs contenthub.Fs) (contenthub.ContentHub, error) {
	log := loggers.NewDefault()

	cs, err := newContentSpec()
	if err != nil {
		return nil, err
	}

	ch := &entity.ContentHub{
		Fs:               fs,
		TemplateExecutor: nil,
		PageCollections: &entity.PageCollections{
			PageMap: &entity.PageMap{
				ContentMap:  newContentMap(),
				ContentSpec: cs,

				Log: log,
			},
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

func newContentMap() *entity.ContentMap {
	m := &entity.ContentMap{
		Pages:    &entity.ContentTree{Name: "pages", Tree: radixtree.New()},
		Sections: &entity.ContentTree{Name: "sections", Tree: radixtree.New()},
	}

	m.PageTrees = []*entity.ContentTree{
		m.Pages, m.Sections,
	}

	m.BundleTrees = []*entity.ContentTree{
		m.Pages, m.Sections,
	}

	return m
}
