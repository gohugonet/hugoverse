package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/mitchellh/mapstructure"
)

const (
	servicesConfigKey = "services"
	privacyConfigKey  = "privacy"

	disqusShortnameKey = "disqusshortname"
	googleAnalyticsKey = "googleanalytics"
	rssLimitKey        = "rssLimit"
)

type ServiceConfig struct {
	Disqus          Disqus
	GoogleAnalytics GoogleAnalytics
	Instagram       Instagram
	Twitter         Twitter
	Vimeo           Vimeo
	YouTube         YouTube
	RSS             RSS
}

// Service is the common values for a service in a policy definition.
type privacyService struct {
	Disable bool
}

// Disqus holds the privacy configuration settings related to the Disqus template.
type Disqus struct {
	privacyService `mapstructure:",squash"`

	// A Shortname is the unique identifier assigned to a Disqus site.
	Shortname string
}

// GoogleAnalytics holds the privacy configuration settings related to the Google Analytics template.
type GoogleAnalytics struct {
	privacyService `mapstructure:",squash"`

	// Enabling this will make the GA templates respect the
	// "Do Not Track" HTTP header. See  https://www.paulfurley.com/google-analytics-dnt/.
	RespectDoNotTrack bool

	// The GA tracking ID.
	ID string
}

// Instagram holds the privacy configuration settings related to the Instagram shortcode.
type Instagram struct {
	privacyService `mapstructure:",squash"`

	// If simple mode is enabled, a static and no-JS version of the Instagram
	// image card will be built.
	Simple bool

	// The Simple variant of the Instagram is decorated with Bootstrap 4 card classes.
	// This means that if you use Bootstrap 4 or want to provide your own CSS, you want
	// to disable the inline CSS provided by Hugo.
	DisableInlineCSS bool

	// App or Client Access Token.
	// If you are using a Client Access Token, remember that you must combine it with your App ID
	// using a pipe symbol (<APPID>|<CLIENTTOKEN>) otherwise the request will fail.
	AccessToken string
}

// Twitter holds the privacy configuration settingsrelated to the Twitter shortcode.
type Twitter struct {
	privacyService `mapstructure:",squash"`

	// When set to true, the Tweet and its embedded page on your site are not used
	// for purposes that include personalized suggestions and personalized ads.
	EnableDNT bool

	// If simple mode is enabled, a static and no-JS version of the Tweet will be built.
	Simple bool

	// The Simple variant of Twitter is decorated with a basic set of inline styles.
	// This means that if you want to provide your own CSS, you want
	// to disable the inline CSS provided by Hugo.
	DisableInlineCSS bool
}

// Vimeo holds the privacy configuration settingsrelated to the Vimeo shortcode.
type Vimeo struct {
	privacyService `mapstructure:",squash"`

	// When set to true, the Vimeo player will be blocked from tracking any session data,
	// including all cookies and stats.
	EnableDNT bool

	// If simple mode is enabled, only a thumbnail is fetched from i.vimeocdn.com and
	// shown with a play button overlaid. If a user clicks the button, he/she will
	// be taken to the video page on vimeo.com in a new browser tab.
	Simple bool
}

// YouTube holds the privacy configuration settingsrelated to the YouTube shortcode.
type YouTube struct {
	privacyService `mapstructure:",squash"`

	// When you turn on privacy-enhanced mode,
	// YouTube won’t store information about visitors on your website
	// unless the user plays the embedded video.
	PrivacyEnhanced bool
}

// RSS holds the functional configuration settings related to the RSS feeds.
type RSS struct {
	// Limit the number of pages.
	Limit int
}

// DecodeServiceConfig creates a services Config from a given Hugo configuration.
func DecodeServiceConfig(cfg config.Provider) (c ServiceConfig, err error) {
	if cfg.IsSet(privacyConfigKey) {
		m := cfg.GetStringMap(privacyConfigKey)

		err = mapstructure.WeakDecode(m, &c)
	}

	if cfg.IsSet(servicesConfigKey) {
		m := cfg.GetStringMap(servicesConfigKey)

		err = mapstructure.WeakDecode(m, &c)
	}

	// Keep backwards compatibility.
	if c.GoogleAnalytics.ID == "" {
		// Try the global config
		c.GoogleAnalytics.ID = cfg.GetString(googleAnalyticsKey)
	}
	if c.Disqus.Shortname == "" {
		c.Disqus.Shortname = cfg.GetString(disqusShortnameKey)
	}

	if c.RSS.Limit == 0 {
		c.RSS.Limit = cfg.GetInt(rssLimitKey)
	}

	return
}
