package module

import "github.com/spf13/afero"

// Module folder structure
const (
	ComponentFolderArchetypes = "archetypes"
	ComponentFolderStatic     = "static"
	ComponentFolderLayouts    = "layouts"
	ComponentFolderContent    = "content"
	ComponentFolderData       = "data"
	ComponentFolderAssets     = "assets"
	ComponentFolderI18n       = "i18n"

	FolderResources = "resources"
)

var (
	ComponentFolders = []string{
		ComponentFolderArchetypes,
		ComponentFolderStatic,
		ComponentFolderLayouts,
		ComponentFolderContent,
		ComponentFolderData,
		ComponentFolderAssets,
		ComponentFolderI18n,
	}
)

type Modules interface {
	Proj() Module
	All() []Module
	IsProjMod(mod Module) bool

	GetSourceLang(source string) (string, bool)
}

type Module interface {
	Owner() Module
	Mounts() []Mount
	Dir() string
}

type Mount interface {
	Source() string
	Target() string
	Lang() string
	Marshal() string
}

type LoadInfo interface {
	Workspace
	Paths
}

type Workspace interface {
	Fs() afero.Fs
	WorkingDir() string
	ThemesDir() string

	DefaultLanguage() string
	OtherLanguageKeys() []string
	GetRelDir(name string, langKey string) (dir string, err error)
}

type Component interface {
	Name() string
	Dir() string
	Language() string
}

type Paths interface {
	ImportPaths() []string
	GetImports(moduleDir string) ([]string, error)
}
