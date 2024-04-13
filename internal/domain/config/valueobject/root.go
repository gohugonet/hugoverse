package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/mitchellh/mapstructure"
)

// RootConfig holds all the top-level configuration options in Hugo
type RootConfig struct {
	// The base URL of the site.
	// Note that the default value is empty, but Hugo requires a valid URL (e.g. "https://example.com/") to work properly.
	// <docsmeta>{"identifiers": ["URL"] }</docsmeta>
	BaseURL string

	// Whether to build content marked as draft.X
	// <docsmeta>{"identifiers": ["draft"] }</docsmeta>
	BuildDrafts bool

	// Whether to build content with expiryDate in the past.
	// <docsmeta>{"identifiers": ["expiryDate"] }</docsmeta>
	BuildExpired bool

	// Whether to build content with publishDate in the future.
	// <docsmeta>{"identifiers": ["publishDate"] }</docsmeta>
	BuildFuture bool

	// Copyright information.
	Copyright string

	// The language to apply to content without any language indicator.
	DefaultContentLanguage string

	// By default, we put the default content language in the root and the others below their language ID, e.g. /no/.
	// Set this to true to put all languages below their language ID.
	DefaultContentLanguageInSubdir bool

	// Disable creation of alias redirect pages.
	DisableAliases bool

	// Disable lower casing of path segments.
	DisablePathToLower bool

	// Disable page kinds from build.
	DisableKinds []string

	// A list of languages to disable.
	DisableLanguages []string

	// The named segments to render.
	// This needs to match the name of the segment in the segments configuration.
	RenderSegments []string

	// Disable the injection of the Hugo generator tag on the home page.
	DisableHugoGeneratorInject bool

	// Disable live reloading in server mode.
	DisableLiveReload bool

	// Enable replacement in Pages' Content of Emoji shortcodes with their equivalent Unicode characters.
	// <docsmeta>{"identifiers": ["Content", "Unicode"] }</docsmeta>
	EnableEmoji bool

	// THe main section(s) of the site.
	// If not set, Hugo will try to guess this from the content.
	MainSections []string

	// Enable robots.txt generation.
	EnableRobotsTXT bool

	// When enabled, Hugo will apply Git version information to each Page if possible, which
	// can be used to keep lastUpdated in synch and to print version information.
	// <docsmeta>{"identifiers": ["Page"] }</docsmeta>
	EnableGitInfo bool

	// Enable to track, calculate and print metrics.
	TemplateMetrics bool

	// Enable to track, print and calculate metric hints.
	TemplateMetricsHints bool

	// Enable to disable the build lock file.
	NoBuildLock bool

	// A list of log IDs to ignore.
	IgnoreLogs []string

	// A list of regexps that match paths to ignore.
	// Deprecated: Use the settings on module imports.
	IgnoreFiles []string

	// Ignore cache.
	IgnoreCache bool

	// Enable to print greppable placeholders (on the form "[i18n] TRANSLATIONID") for missing translation strings.
	EnableMissingTranslationPlaceholders bool

	// Enable to panic on warning log entries. This may make it easier to detect the source.
	PanicOnWarning bool

	// The configured environment. Default is "development" for server and "production" for build.
	Environment string

	// The default language code.
	LanguageCode string

	// Enable if the site content has CJK language (Chinese, Japanese, or Korean). This affects how Hugo counts words.
	HasCJKLanguage bool

	// The default number of pages per page when paginating.
	Paginate int

	// The path to use when creating pagination URLs, e.g. "page" in /page/2/.
	PaginatePath string

	// Whether to pluralize default list titles.
	// Note that this currently only works for English, but you can provide your own title in the content file's front matter.
	PluralizeListTitles bool

	// Whether to capitalize automatic page titles, applicable to section, taxonomy, and term pages.
	CapitalizeListTitles bool

	// Make all relative URLs absolute using the baseURL.
	// <docsmeta>{"identifiers": ["baseURL"] }</docsmeta>
	CanonifyURLs bool

	// Enable this to make all relative URLs relative to content root. Note that this does not affect absolute URLs.
	RelativeURLs bool

	// Removes non-spacing marks from composite characters in content paths.
	RemovePathAccents bool

	// Whether to track and print unused templates during the build.
	PrintUnusedTemplates bool

	// Enable to print warnings for missing translation strings.
	PrintI18nWarnings bool

	// ENable to print warnings for multiple files published to the same destination.
	PrintPathWarnings bool

	// URL to be used as a placeholder when a page reference cannot be found in ref or relref. Is used as-is.
	RefLinksNotFoundURL string

	// When using ref or relref to resolve page links and a link cannot be resolved, it will be logged with this log level.
	// Valid values are ERROR (default) or WARNING. Any ERROR will fail the build (exit -1).
	RefLinksErrorLevel string

	// This will create a menu with all the sections as menu items and all the sections’ pages as “shadow-members”.
	SectionPagesMenu string

	// The length of text in words to show in a .Summary.
	SummaryLength int

	// The site title.
	Title string

	// The theme(s) to use.
	// See Modules for more a more flexible way to load themes.
	Theme []string

	// Timeout for generating page contents, specified as a duration or in seconds.
	Timeout string

	// The time zone (or location), e.g. Europe/Oslo, used to parse front matter dates without such information and in the time function.
	TimeZone string

	// Set titleCaseStyle to specify the title style used by the title template function and the automatic section titles in Hugo.
	// It defaults to AP Stylebook for title casing, but you can also set it to Chicago or Go (every word starts with a capital letter).
	TitleCaseStyle string

	// The editor used for opening up new content.
	NewContentEditor string

	// Don't sync modification time of files for the static mounts.
	NoTimes bool

	// Don't sync modification time of files for the static mounts.
	NoChmod bool

	// Clean the destination folder before a new build.
	// This currently only handles static files.
	CleanDestinationDir bool

	// A Glob pattern of module paths to ignore in the _vendor folder.
	IgnoreVendorPaths string

	CommonDirs `mapstructure:",squash"`
}

func DecodeRoot(provider config.Provider) (RootConfig, error) {
	conf := RootConfig{}

	if err := mapstructure.WeakDecode(provider.Get(""), &conf); err != nil {
		return RootConfig{}, err
	}

	return conf, nil
}
