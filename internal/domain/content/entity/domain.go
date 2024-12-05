package entity

import (
	"encoding/json"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
)

func (c *Content) ApplyDomain(siteId string, domain string) (*valueobject.Domain, error) {
	site, err := c.getContent("Site", siteId)
	if err != nil {
		return nil, err
	}

	if site, ok := site.(*valueobject.Site); ok {
		slug, err := Slug(site)
		if err != nil {
			return nil, err
		}

		sd, err := c.searchDomain(domain, slug, site.Owner)
		if err != nil {
			return nil, err
		}

		if sd == nil {
			item, err := valueobject.NewItemWithNamespace("Domain")
			if err != nil {
				return nil, err
			}

			sd = &valueobject.Domain{
				Item:  *item,
				Root:  domain,
				Sub:   slug,
				Owner: site.Owner,
			}

			_, err = c.newContent("Domain", sd)
			if err != nil {
				return nil, err
			}
		}

		return sd, nil
	}

	return nil, errors.New("only site could be deployed with domain")
}

func (c *Content) searchDomain(root string, sub string, owner string) (*valueobject.Domain, error) {
	// 构建精确匹配的查询条件
	conditions := map[string]string{
		"root":  root,
		"sub":   sub,
		"owner": owner,
	}

	// 查询域名信息
	domains, err := c.termSearch("Domain", conditions)
	if err != nil {
		return nil, err
	}

	// 如果返回仅一个结果且为 nil，返回 nil
	if len(domains) == 1 && domains[0] == nil {
		return nil, nil
	}

	// 遍历查询结果并解析
	for _, data := range domains {
		var domain valueobject.Domain
		if err := json.Unmarshal(data, &domain); err != nil {
			return nil, err
		}

		// 返回匹配的域名对象
		return &domain, nil
	}

	// 未找到匹配结果
	return nil, nil
}
