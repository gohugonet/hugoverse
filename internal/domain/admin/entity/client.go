package entity

import "github.com/gohugonet/hugoverse/internal/domain/admin/valueobject"

type Client struct {
	Conf *valueobject.Config
}

func (a *Client) ClientSecret() string { return a.Conf.ClientSecret }
