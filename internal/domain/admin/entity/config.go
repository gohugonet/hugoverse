package entity

import contentEntity "github.com/gohugonet/hugoverse/internal/domain/content/entity"

type Config struct {
	contentEntity.Item

	Name                    string   `json:"name"`
	Domain                  string   `json:"domain"`
	BindAddress             string   `json:"bind_addr"`
	HTTPPort                int      `json:"http_port"`
	HTTPSPort               int      `json:"https_port"`
	AdminEmail              string   `json:"admin_email"`
	ClientSecret            string   `json:"client_secret"`
	Etag                    string   `json:"etag"`
	DisableCORS             bool     `json:"cors_disabled"`
	DisableGZIP             bool     `json:"gzip_disabled"`
	DisableHTTPCache        bool     `json:"cache_disabled"`
	CacheMaxAge             int64    `json:"cache_max_age"`
	CacheInvalidate         []string `json:"cache"`
	BackupBasicAuthUser     string   `json:"backup_basic_auth_user"`
	BackupBasicAuthPassword string   `json:"backup_basic_auth_password"`
}

func (c *Config) isCacheInvalidate() bool {
	return len(c.CacheInvalidate) > 0 && c.CacheInvalidate[0] == "invalidate"
}
