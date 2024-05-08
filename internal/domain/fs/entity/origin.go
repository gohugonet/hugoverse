package entity

import "github.com/spf13/afero"

type OriginFs struct {
	// Source is Hugo's source file system.
	// Note that this will always be a "plain" Afero filesystem:
	// * afero.OsFs when running in production
	// * afero.MemMapFs for many of the tests.
	Source afero.Fs

	// PublishDir is where Hugo publishes its rendered content.
	// It's mounted inside publishDir (default /public).
	PublishDir afero.Fs

	// WorkingDirReadOnly is a read-only file system
	// restricted to the project working dir.
	WorkingDirReadOnly afero.Fs
}

func (f *OriginFs) Origin() afero.Fs {
	return f.Source
}

func (f *OriginFs) SourceFs() afero.Fs {
	return f.Origin()
}

func (f *OriginFs) Publish() afero.Fs {
	return f.PublishDir
}

func (f *OriginFs) PublishFs() afero.Fs {
	return f.Publish()
}

func (f *OriginFs) Working() afero.Fs {
	return f.WorkingDirReadOnly
}

func getWorkingDirFsReadOnly(base afero.Fs, workingDir string) afero.Fs {
	if workingDir == "" {
		return afero.NewReadOnlyFs(base)
	}
	return afero.NewBasePathFs(afero.NewReadOnlyFs(base), workingDir)
}
