package valueobject

import (
	"github.com/mdfriday/hugoverse/internal/domain/config"
	"github.com/mdfriday/hugoverse/pkg/hexec"
	"github.com/mdfriday/hugoverse/pkg/hexec/security"
	"github.com/mdfriday/hugoverse/pkg/types"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type SecurityConfig struct {
	// Restricts access to os.Exec....
	// <docsmeta>{ "newIn": "0.91.0" }</docsmeta>
	hexec.Auth `json:"exec"`

	// Restricts access to certain template funcs.
	Funcs Funcs `json:"funcs"`

	// Restricts access to resources.GetRemote, getJSON, getCSV.
	HTTP HTTP `json:"http"`

	// Allow inline shortcodes
	EnableInlineShortcodes bool `json:"enableInlineShortcodes"`
}

// Funcs holds template funcs policies.
type Funcs struct {
	// OS env keys allowed to query in os.Getenv.
	Getenv security.Whitelist `json:"getenv"`
}

type HTTP struct {
	// URLs to allow in remote HTTP (resources.Get, getJSON, getCSV).
	URLs security.Whitelist `json:"urls"`

	// HTTP methods to allow.
	Methods security.Whitelist `json:"methods"`

	// Media types where the Content-Type in the response is used instead of resolving from the file content.
	MediaTypes security.Whitelist `json:"mediaTypes"`
}

const securityConfigKey = "security"

// DefaultSecurityConfig holds the default security policy.
var DefaultSecurityConfig = SecurityConfig{
	Auth: hexec.Auth{
		Allow: security.MustNewWhitelist(
			"^(dart-)?sass(-embedded)?$", // sass, dart-sass, dart-sass-embedded.
			"^go$",                       // for Go Modules
			"^npx$",                      // used by all Node tools (Babel, PostCSS).
			"^postcss$",
		),
		// These have been tested to work with Hugo's external programs
		// on Windows, Linux and MacOS.
		OsEnv: security.MustNewWhitelist(`(?i)^((HTTPS?|NO)_PROXY|PATH(EXT)?|APPDATA|TE?MP|TERM|GO\w+|(XDG_CONFIG_)?HOME|USERPROFILE|SSH_AUTH_SOCK|DISPLAY|LANG|SYSTEMDRIVE)$`),
	},
	Funcs: Funcs{
		Getenv: security.MustNewWhitelist("^HUGO_", "^CI$"),
	},
	HTTP: HTTP{
		URLs:    security.MustNewWhitelist(".*"),
		Methods: security.MustNewWhitelist("(?i)GET|POST"),
	},
}

func DecodeSecurityConfig(cfg config.Provider) (SecurityConfig, error) {
	sc := DefaultSecurityConfig
	if cfg.IsSet(securityConfigKey) {
		m := cfg.GetStringMap(securityConfigKey)
		dec, err := mapstructure.NewDecoder(
			&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           &sc,
				DecodeHook:       stringSliceToWhitelistHook(),
			},
		)
		if err != nil {
			return sc, err
		}

		if err = dec.Decode(m); err != nil {
			return sc, err
		}
	}

	if !sc.EnableInlineShortcodes {
		// Legacy
		sc.EnableInlineShortcodes = cfg.GetBool("enableInlineShortcodes")
	}

	return sc, nil
}

func stringSliceToWhitelistHook() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if t != reflect.TypeOf(security.Whitelist{}) {
			return data, nil
		}

		wl := types.ToStringSlicePreserveString(data)

		return security.NewWhitelist(wl...)
	}
}
