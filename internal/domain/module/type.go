package module

// Module folder structure
const (
	ComponentFolderArchetypes = "archetypes"
	ComponentFolderStatic     = "static"
	ComponentFolderLayouts    = "layouts"
	ComponentFolderContent    = "content"
	ComponentFolderData       = "data"
	ComponentFolderAssets     = "assets"
	ComponentFolderI18n       = "i18n"
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
}

type Module interface {
	// Config The decoded module config and mounts.
	Config() ModuleConfig
	// Owner In the dependency tree, this is the first module that defines this module
	// as a dependency.
	Owner() Module
	// Mounts Any directory remappings.
	Mounts() []Mount

	IsProj() bool
}

// ModuleConfig holds a module config.
type ModuleConfig struct {
	Mounts  []Mount
	Imports []Import
}

type Mount struct {
	// relative pathspec in source repo, e.g. "scss"
	Source string
	// relative target pathspec, e.g. "assets/bootstrap/scss"
	Target string
	// any language code associated with this mount.
	Lang string
}

type Import struct {
	// Module pathspec
	Path string
}
