package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/lazy"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type Page struct {
	*Source
	*FrontMatter
	*Path
	*Shortcodes
	*Content

	kind     string
	singular string
	term     string

	taxonomyService contenthub.TaxonomyService

	bundled bool // Set if this page is bundled inside another.
}

func newBundledPage(source *Source, langSer contenthub.LangService, taxSer contenthub.TaxonomyService, tmplSvc contenthub.Template) (*Page, error) {
	contentBytes, err := source.contentSource()
	if err != nil {
		return nil, err
	}

	p := &Page{
		Source: source,
		FrontMatter: &FrontMatter{
			Params:     maps.Params{},
			Customized: maps.Params{},

			langService: langSer,
		},
		Shortcodes: &Shortcodes{source: contentBytes, ordinal: 0, tmplSvc: tmplSvc, pid: pid},
		Content:    &Content{source: contentBytes},

		bundled: true,

		taxonomyService: taxSer,
	}

	p.Source.registerHandler(p.FrontMatter.frontMatterHandler,
		p.Content.summaryHandler, p.Content.bytesHandler,
		p.Shortcodes.shortcodeHandler)

	if err := p.Source.parse(); err != nil {
		return nil, err
	}

	if err := p.FrontMatter.parse(); err != nil {
		return nil, err
	}

	p.setupPagePath()
	p.setupLang()
	p.setupKind()

	return p, nil
}

func (p *Page) setupPagePath() {
	pi := paths.Parse(p.Source.fi.Component(), p.Source.fi.FileName())
	if p.FrontMatter.Path != "" {
		p.Path = newPathFromConfig(p.FrontMatter.Path, p.FrontMatter.Kind, pi)
	} else {
		p.Path = &Path{pathInfo: pi}
	}
}

func (p *Page) setupLang() {
	l, ok := p.FrontMatter.langService.GetSourceLang(p.Source.fi.Root())
	if ok {
		idx, err := p.FrontMatter.langService.GetLanguageIndex(l)
		if err != nil {
			panic(err)
		}
		p.Identity.Lang = l
		p.Identity.LangIdx = idx
	} else {
		panic(fmt.Sprintf("unknown lang %q", p.Source.fi.Root()))
	}
}

func (p *Page) setupKind() {
	p.kind = p.FrontMatter.Kind
	if p.FrontMatter.Kind == "" {
		p.Kind = valueobject.KindSection
		if p.Path.pathInfo.Base() == "/" {
			p.Kind = valueobject.KindHome
		} else if p.Path.pathInfo.IsBranchBundle() {
			// A section, taxonomy or term.
			if !p.taxonomyService.IsZero(p.Path.Path()) {
				// Either a taxonomy or a term.
				if p.taxonomyService.PluralTreeKey(p.Path.Path()) == p.Path.Path() {
					p.Kind = valueobject.KindTaxonomy
					p.singular = p.taxonomyService.Singular(p.Path.Path())
				} else {
					p.Kind = valueobject.KindTerm
					p.singular = p.taxonomyService.Singular(p.Path.Path())
					p.term = p.Path.pathInfo.Unnormalized().BaseNameNoIdentifier()
				}
			}
		} else {
			p.Kind = valueobject.KindPage
		}
	}
}

func newPage(m *pageMeta) (*pageState, *paths.Path, error) {
	m.Staler = &stale.AtomicStaler{}

	if m.pageConfig == nil {
		m.pageMetaParams = pageMetaParams{
			pageConfig: &pagemeta.PageConfig{
				Params: maps.Params{},
			},
		}
	}

	var sourceKey string
	if m.f != nil {
		sourceKey = filepath.ToSlash(m.f.Filename())
	}

	pid := pageIDCounter.Add(1)
	pi, err := m.parseFrontMatter(h, pid, sourceKey)
	if err != nil {
		return nil, nil, err
	}

	if err := m.setMetaPre(pi, h.Log, h.Conf); err != nil {
		return nil, nil, m.wrapError(err, h.BaseFs.SourceFs)
	}
	pcfg := m.pageConfig
	if pcfg.Lang != "" {
		if h.Conf.IsLangDisabled(pcfg.Lang) {
			return nil, nil, nil
		}
	}

	if pcfg.Path != "" {
		s := m.pageConfig.Path
		if !paths.HasExt(s) {
			var (
				isBranch bool
				ext      string = "md"
			)
			if pcfg.Kind != "" {
				isBranch = kinds.IsBranch(pcfg.Kind)
			} else if m.pathInfo != nil {
				isBranch = m.pathInfo.IsBranchBundle()
				if m.pathInfo.Ext() != "" {
					ext = m.pathInfo.Ext()
				}
			} else if m.f != nil {
				pi := m.f.FileInfo().Meta().PathInfo
				isBranch = pi.IsBranchBundle()
				if pi.Ext() != "" {
					ext = pi.Ext()
				}
			}
			if isBranch {
				s += "/_index." + ext
			} else {
				s += "/index." + ext
			}
		}
		m.pathInfo = h.Conf.PathParser().Parse(files.ComponentFolderContent, s)
	} else if m.pathInfo == nil {
		if m.f != nil {
			m.pathInfo = m.f.FileInfo().Meta().PathInfo
		}

		if m.pathInfo == nil {
			panic(fmt.Sprintf("missing pathInfo in %v", m))
		}
	}

	ps, err := func() (*pageState, error) {
		if m.s == nil {
			// Identify the Site/language to associate this Page with.
			var lang string
			if pcfg.Lang != "" {
				lang = pcfg.Lang
			} else if m.f != nil {
				meta := m.f.FileInfo().Meta()
				lang = meta.Lang
				m.s = h.Sites[meta.LangIndex]
			} else {
				lang = m.pathInfo.Lang()
			}
			if lang == "" {
				lang = h.Conf.DefaultContentLanguage()
			}
			var found bool
			for _, ss := range h.Sites {
				if ss.Lang() == lang {
					m.s = ss
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("no site found for language %q", lang)
			}
		}

		// Identify Page Kind.
		if m.pageConfig.Kind == "" {
			m.pageConfig.Kind = kinds.KindSection
			if m.pathInfo.Base() == "/" {
				m.pageConfig.Kind = kinds.KindHome
			} else if m.pathInfo.IsBranchBundle() {
				// A section, taxonomy or term.
				tc := m.s.pageMap.cfg.getTaxonomyConfig(m.Path())
				if !tc.IsZero() {
					// Either a taxonomy or a term.
					if tc.pluralTreeKey == m.Path() {
						m.pageConfig.Kind = kinds.KindTaxonomy
						m.singular = tc.singular
					} else {
						m.pageConfig.Kind = kinds.KindTerm
						m.term = m.pathInfo.Unnormalized().BaseNameNoIdentifier()
						m.singular = tc.singular
					}
				}
			} else if m.f != nil {
				m.pageConfig.Kind = kinds.KindPage
			}
		}

		if m.pageConfig.Kind == kinds.KindPage && !m.s.conf.IsKindEnabled(m.pageConfig.Kind) {
			return nil, nil
		}

		// Parse the rest of the page content.
		m.content, err = m.newCachedContent(h, pi)
		if err != nil {
			return nil, m.wrapError(err, h.SourceFs)
		}

		ps := &pageState{
			pid:                               pid,
			pageOutput:                        nopPageOutput,
			pageOutputTemplateVariationsState: &atomic.Uint32{},
			resourcesPublishInit:              &sync.Once{},
			Staler:                            m,
			dependencyManager:                 m.s.Conf.NewIdentityManager(m.Path()),
			pageCommon: &pageCommon{
				FileProvider:              m,
				AuthorProvider:            m,
				Scratcher:                 maps.NewScratcher(),
				store:                     maps.NewScratch(),
				Positioner:                page.NopPage,
				InSectionPositioner:       page.NopPage,
				ResourceNameTitleProvider: m,
				ResourceParamsProvider:    m,
				PageMetaProvider:          m,
				RelatedKeywordsProvider:   m,
				OutputFormatsProvider:     page.NopPage,
				ResourceTypeProvider:      pageTypesProvider,
				MediaTypeProvider:         pageTypesProvider,
				RefProvider:               page.NopPage,
				ShortcodeInfoProvider:     page.NopPage,
				LanguageProvider:          m.s,

				InternalDependencies: m.s,
				init:                 lazy.New(),
				m:                    m,
				s:                    m.s,
				sWrapped:             page.WrapSite(m.s),
			},
		}

		if m.f != nil {
			gi, err := m.s.h.gitInfoForPage(ps)
			if err != nil {
				return nil, fmt.Errorf("failed to load Git data: %w", err)
			}
			ps.gitInfo = gi
			owners, err := m.s.h.codeownersForPage(ps)
			if err != nil {
				return nil, fmt.Errorf("failed to load CODEOWNERS: %w", err)
			}
			ps.codeowners = owners
		}

		ps.pageMenus = &pageMenus{p: ps}
		ps.PageMenusProvider = ps.pageMenus
		ps.GetPageProvider = pageSiteAdapter{s: m.s, p: ps}
		ps.GitInfoProvider = ps
		ps.TranslationsProvider = ps
		ps.ResourceDataProvider = &pageData{pageState: ps}
		ps.RawContentProvider = ps
		ps.ChildCareProvider = ps
		ps.TreeProvider = pageTree{p: ps}
		ps.Eqer = ps
		ps.TranslationKeyProvider = ps
		ps.ShortcodeInfoProvider = ps
		ps.AlternativeOutputFormatsProvider = ps

		if err := ps.initLazyProviders(); err != nil {
			return nil, ps.wrapError(err)
		}
		return ps, nil
	}()
	// Make sure to evict any cached and now stale data.
	if err != nil {
		m.MarkStale()
	}

	if ps == nil {
		return nil, nil, err
	}

	return ps, ps.PathInfo(), err
}
