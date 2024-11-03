package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"net/url"
)

func (c *Content) GetDeployment(siteId string) (*valueobject.SiteDeployment, error) {
	content, err := c.getContent("Site", siteId)
	if err != nil {
		return nil, err
	}

	if site, ok := content.(*valueobject.Site); ok {
		sd, err := c.searchDeployment(siteId)
		if err != nil {
			return nil, err
		}

		if sd == nil {
			item, err := valueobject.NewItemWithNamespace("SiteDeployment")
			if err != nil {
				return nil, err
			}

			sd = &valueobject.SiteDeployment{
				Item:    *item,
				Site:    site.QueryString(),
				Netlify: fmt.Sprintf("mdf-%s", site.ItemSlug()),
				Domain:  site.ItemSlug(),
				Status:  "Not Started",
			}

			_, err = c.newContent("SiteDeployment", sd)
			if err != nil {
				return nil, err
			}
		}

		return sd, nil
	}

	return nil, errors.New("only site could be deployed")
}

func (c *Content) searchDeployment(siteId string) (*valueobject.SiteDeployment, error) {
	q := fmt.Sprintf(`site:"api/content?type=Site&id=%s"`, siteId)
	encodedQ := url.QueryEscape(q)

	siteDeployments, err := c.search("SiteDeployment", fmt.Sprintf("site:%s", encodedQ))
	if err != nil {
		return nil, err
	}

	for _, data := range siteDeployments {
		var sd valueobject.SiteDeployment
		if err := json.Unmarshal(data, &sd); err != nil {
			return nil, err
		}

		return &sd, nil
	}

	return nil, nil
}
