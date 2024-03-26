package valueobject

// BaseFs contains the core base filesystems used by Hugo. The name "base" is used
// to underline that even if they can be composites, they all have a base path set to a specific
// resource folder, e.g "/my-project/content". So, no absolute filenames needed.
type BaseFs struct {

	// SourceFilesystems contains the different source file systems.
	*SourceFilesystems

	//// The project source.
	//SourceFs afero.Fs
	//
	//// The filesystem used to publish the rendered site.
	//// This usually maps to /my-project/public.
	//PublishFs afero.Fs
	//
	//// A read-only filesystem starting from the project workDir.
	//WorkDir afero.Fs

	TheBigFs *FilesystemsCollector
}
