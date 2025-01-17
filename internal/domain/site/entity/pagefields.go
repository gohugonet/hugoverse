package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"time"
)

func (p *Page) Section() string {
	return p.Page.Section()
}

func (p *Page) Params() maps.Params {
	return p.Page.Params()
}

func (p *Page) Resources() PageResources {
	return p.resources
}

func (p *Page) Date() time.Time {
	return p.Page.PageDate()
}

func (p *Page) PublishDate() time.Time {
	return time.Now()
}

func (p *Page) Lastmod() time.Time {
	return time.Now()
}

func (p *Page) ExpiryDate() time.Time {
	return time.Now().AddDate(1, 0, 0)
}

func (p *Page) File() contenthub.File {
	return p.Page.PageFile()
}

func (p *Page) OutputFormats() valueobject.OutputFormats {
	// TODO
	return make(valueobject.OutputFormats, 0)
}

func (p *Page) Data() any {
	p.dataInit.Do(func() {
		p.data = make(Data)

		if p.Kind() == contenthub.KindPage {
			return
		}

		p.data["pages"] = p.Pages
	})

	return p.data
}

func (p *Page) Pages() []*Page {
	ps := p.Page.Pages(p.Site.Language.CurrentLanguageIndex())
	if ps == nil {
		return make([]*Page, 0)
	}

	return p.sitePages(ps)
}

func (p *Page) GetTerms(taxonomy string) []*Page {
	ps := p.Page.Terms(p.Site.Language.CurrentLanguageIndex(), taxonomy)
	if ps == nil {
		return make([]*Page, 0)
	}

	return p.sitePages(ps)
}

func (p *Page) Translations() []*Page {
	return p.sitePages(p.Page.Translations())
}

func (p *Page) Parent() *Page {
	if p.IsHome() {
		return nil
	}

	page := p.Page.Parent()
	if page == nil {
		return p.Site.home
	}

	sp, err := p.sitePage(page)
	if err != nil {
		return nil
	}

	return sp
}

func (p *Page) getPageOutput(chp contenthub.Page) contenthub.PageOutput {
	pos, err := chp.PageOutputs()
	if err != nil {
		p.Log.Errorln("getPageOutput", err)
		panic(err)
	}

	for _, po := range pos {
		if po.TargetFormat().MediaType == p.PageOutput.TargetFormat().MediaType {
			return po
		}
	}

	p.Log.Errorln("getPageOutput", "no page output")
	panic("no page output")
}

func (p *Page) Content() (any, error) {
	return p.PageOutput.Content()
}

func (p *Page) Plain() string {
	// TODO
	return p.Page.RawContent()
}

func (p *Page) IsAncestor(other any) bool {
	op, ok := other.(*Page)
	if !ok {
		return false
	}

	return p.Page.IsAncestor(op.Page)
}

func (p *Page) GitInfo() valueobject.GitInfo {
	if p.git == nil {
		return valueobject.GitInfo{}
	}

	return p.git.GetInfo(p.Page.PageFile().Filename())
}

func (p *Page) Title() string {
	return p.Page.Title()
}

func (p *Page) Language() struct {
	Lang         string
	LanguageName string
	LanguageCode string
} {
	return struct {
		Lang         string
		LanguageName string
		LanguageCode string
	}{
		Lang:         p.PageIdentity().PageLanguage(),
		LanguageName: p.Site.Language.LangSvc.GetLanguageName(p.PageIdentity().PageLanguage()),
		LanguageCode: p.PageIdentity().PageLanguage(),
	}
}

func (p *Page) Description() string {
	return p.Page.Description()
}
