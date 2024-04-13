package valueobject

type BaseDirs struct {
	WorkingDir string
	ThemesDir  string
	PublishDir string
}

type CommonDirs struct {
	// The directory where Hugo will look for themes.
	ThemesDir string

	// Where to put the generated files.
	PublishDir string

	// The directory to put the generated resources files. This directory should in most situations be considered temporary
	// and not be committed to version control. But there may be cached content in here that you want to keep,
	// e.g. resources/_gen/images for performance reasons or CSS built from SASS when your CI server doesn't have the full setup.
	ResourceDir string

	// The project root directory.
	WorkingDir string

	// The root directory for all cache files.
	CacheDir string

	// The content source directory.
	// Deprecated: Use module mounts.
	ContentDir string
	// Deprecated: Use module mounts.
	// The data source directory.
	DataDir string
	// Deprecated: Use module mounts.
	// The layout source directory.
	LayoutDir string
	// Deprecated: Use module mounts.
	// The i18n source directory.
	I18nDir string
	// Deprecated: Use module mounts.
	// The archetypes source directory.
	ArcheTypeDir string
	// Deprecated: Use module mounts.
	// The assets source directory.
	AssetDir string
}
