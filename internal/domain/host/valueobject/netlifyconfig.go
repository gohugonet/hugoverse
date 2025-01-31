package valueobject

type NetlifyConfig struct {
	AuthToken     string `envconfig:"auth_token" required:"true"`
	SiteID        string `envconfig:"site_id" required:"true"`
	SiteName      string `default:""`
	FullDomain    string `default:""`
	Directory     string `required:"true"`
	Draft         bool   `default:"true"`
	DeployMessage string `default:""`
}
