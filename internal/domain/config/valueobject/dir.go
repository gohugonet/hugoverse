package valueobject

type BaseDirs struct {
	WorkingDir string
	ThemesDir  string
	PublishDir string
	CacheDir   string
}

type CommonDirs struct {
	BaseDirs `mapstructure:",squash"`

	// The directory to put the generated resources files. This directory should in most situations be considered temporary
	// and not be committed to version control. But there may be cached content in here that you want to keep,
	// e.g. resources/_gen/images for performance reasons or CSS built from SASS when your CI server doesn't have the full setup.
	ResourceDir string

	// The content source directory.
	ContentDir string
	// The data source directory.
	DataDir string
	// The layout source directory.
	LayoutDir string
	// The i18n source directory.
	I18nDir string
	// The archetypes source directory.
	ArcheTypeDir string
	// The assets source directory.
	AssetDir string
}

func (dirs CommonDirs) GetDirectoryByName(folderName string) string {
	switch folderName {
	case "layout":
		return dirs.LayoutDir
	case "content":
		return dirs.ContentDir
	case "data":
		return dirs.DataDir
	case "i18n":
		return dirs.I18nDir
	case "archetypes":
		return dirs.ArcheTypeDir
	case "assets":
		return dirs.AssetDir
	default:
		return ""
	}
}
