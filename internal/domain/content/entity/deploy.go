package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"time"
)

func (c *Content) GetDeployment(siteId string, domain *valueobject.Domain) (*valueobject.SiteDeployment, error) {
	content, err := c.getContent("Site", siteId)
	if err != nil {
		return nil, err
	}

	if site, ok := content.(*valueobject.Site); ok {
		sd, err := c.searchDeployment(domain.Root, domain.Sub, domain.Owner)
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
				Netlify: fmt.Sprintf("mdf-%d-%s", time.Now().UnixMilli(), site.ItemSlug()),
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

func (c *Content) searchDeployment(root string, sub string, owner string) (*valueobject.SiteDeployment, error) {
	conditions := map[string]string{
		"root":   root,
		"domain": sub,
		"owner":  owner,
	}

	// 查询域名信息
	deploys, err := c.termSearch("SiteDeployment", conditions)
	if err != nil {
		return nil, err
	}

	// 如果返回仅一个结果且为 nil，返回 nil
	if len(deploys) == 1 && deploys[0] == nil {
		return nil, nil
	}

	// 遍历查询结果并解析
	for _, data := range deploys {
		var deployment valueobject.SiteDeployment
		if err := json.Unmarshal(data, &deployment); err != nil {
			return nil, err
		}

		// 返回匹配的域名对象
		return &deployment, nil
	}

	// 未找到匹配结果
	return nil, nil
}
