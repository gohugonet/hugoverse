package entity

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"gopkg.in/yaml.v3"
	"path"
	"strconv"
)

type Hugo struct {
	Services content.Services

	contentSvc contentSvc

	site *valueobject.Site

	Log loggers.Logger
}

func (h *Hugo) LoadProject(c contentSvc) error {
	h.contentSvc = c

	if err := h.loadSite(); err != nil {
		return err
	}

	if err := h.loadSiteLanguages(); err != nil {
		return err
	}

	if err := h.loadPosts(); err != nil {
		return err
	}

	return nil
}

func (h *Hugo) loadPosts() error {
	authorQueryStr, err := h.getAuthor()
	if err != nil {
		return err
	}

	codes := h.Services.LanguageKeys()
	for _, code := range codes {
		langIndex, err := h.Services.GetLanguageIndex(code)
		if err != nil {
			return err
		}

		if err := h.Services.WalkPages(langIndex, func(p contenthub.Page) error {
			if p.PageIdentity().PageLanguage() != code {
				return nil
			}
			h.Log.Printf("Loading post: %s, %s-%s\n",
				p.PageFile().FileInfo().RelativeFilename(), code, p.PageIdentity().PageLanguage())

			i, err := valueobject.NewItemWithNamespace("Post")
			if err != nil {
				return err
			}

			post := &valueobject.Post{
				Item:    *i,
				Title:   p.Title(),
				Author:  authorQueryStr,
				Content: p.RawContent(),
			} // TODO, page assets
			id, err := h.contentSvc.newContent("Post", post)
			if err != nil {
				return err
			}

			h.Log.Printf("Loaded post: %+v", post)

			num, err := strconv.Atoi(id)
			if err != nil {
				return err
			}
			post.ID = num

			spi, err := valueobject.NewItemWithNamespace("SitePost")
			if err != nil {
				return err
			}

			sitePost := &valueobject.SitePost{
				Item: *spi,
				Post: post.QueryString(),
				Site: h.site.QueryString(),
				Path: path.Join(p.PageFile().FileInfo().Root(), p.PageFile().FileInfo().RelativeFilename()),
			}
			_, err = h.contentSvc.newContent("SitePost", sitePost)
			if err != nil {
				return err
			}

			h.Log.Printf("Loaded SitePost: %+v", sitePost)

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Hugo) loadSiteLanguages() error {
	codes := h.Services.LanguageKeys()
	for _, code := range codes {
		i, err := valueobject.NewItemWithNamespace("SiteLanguage")
		if err != nil {
			return err
		}

		langQueryStr, err := h.getLanguage(code)
		if err != nil {
			return err
		}
		h.Log.Println("Get language", code, langQueryStr)

		siteLang := &valueobject.SiteLanguage{
			Item:     *i,
			Site:     h.site.QueryString(),
			Language: langQueryStr,
			Default:  code == h.Services.DefaultLanguage(),
			Folder:   h.Services.GetLanguageFolder(code),
		}
		h.Log.Printf("Loadeding SiteLanguage: %+v", *siteLang)
		_, err = h.contentSvc.newContent("SiteLanguage", siteLang)
		if err != nil {
			return err
		}

		h.Log.Printf("Loaded SiteLanguage: %+v", *siteLang)
	}
	return nil
}

func (h *Hugo) loadSite() error {
	i, err := valueobject.NewItemWithNamespace("Site")
	if err != nil {
		return err
	}

	themeQueryStr, err := h.getTheme(h.Services.DefaultTheme())
	if err != nil {
		return err
	}

	site := &valueobject.Site{
		Item:       *i,
		Title:      h.Services.SiteTitle(),
		BaseURL:    h.Services.BaseUrl(),
		WorkingDir: h.Services.WorkingDir(),
		Theme:      themeQueryStr,
		Owner:      1,
	}
	if h.Services.ConfigParams() != nil {
		site.Params, err = mapToYAML(h.Services.ConfigParams())
		if err != nil {
			return err
		}
	}

	id, err := h.contentSvc.newContent("Site", site)
	if err != nil {
		return err
	}

	h.Log.Printf("Loaded Site: %+v", *site)

	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	site.ID = num
	h.site = site

	return nil
}

func (h *Hugo) getTheme(moduleUrl string) (string, error) {
	data, err := h.contentSvc.search("Theme", fmt.Sprintf("module_url:%s", moduleUrl))
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		firstData := data[0]
		var result valueobject.Theme
		if err := json.Unmarshal(firstData, &result); err != nil {
			return "", err
		}

		return result.QueryString(), nil
	}

	return "", fmt.Errorf("no themes found")
}

func (h *Hugo) getLanguage(code string) (string, error) {
	data, err := h.contentSvc.search("Language", fmt.Sprintf("code:%s", code))
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		firstData := data[0]
		var result valueobject.Language
		if err := json.Unmarshal(firstData, &result); err != nil {
			return "", err
		}

		return result.QueryString(), nil
	}

	return "", fmt.Errorf("no languages found")
}

func (h *Hugo) getAuthor() (string, error) {
	//TODO get user email from token

	data, err := h.contentSvc.search("Author", fmt.Sprintf("email:%s", "me@sunwei.xyz"))
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		firstData := data[0]
		var result valueobject.Author
		if err := json.Unmarshal(firstData, &result); err != nil {
			return "", err
		}

		return result.QueryString(), nil
	}

	return "", fmt.Errorf("no authors found")
}

func mapToYAML(data map[string]any) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}
