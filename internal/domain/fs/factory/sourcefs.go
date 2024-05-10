package factory

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

func newSourceFilesystem(name string, fs, sourceFs afero.Fs, dirs []valueobject.FileMetaInfo) *valueobject.SourceFilesystem {
	return &valueobject.SourceFilesystem{
		Name:     name,
		Fs:       fs,
		SourceFs: sourceFs,
		Dirs:     dirs,
	}
}

type sourceFilesystemsBuilder struct {
	modules  module.Modules
	sourceFs afero.Fs
	result   *valueobject.SourceFilesystems
	theBigFs *valueobject.FilesystemsCollector
}

func (b *sourceFilesystemsBuilder) Build() (*valueobject.SourceFilesystems, error) {
	if b.theBigFs == nil {
		// Modules - mounts <-> RootMappingFs - OverlayFS
		theBigFs, err := b.createMainOverlayFs()
		if err != nil {
			return nil, fmt.Errorf("create main fs: %w", err)
		}

		b.theBigFs = theBigFs
	}

	createView := func(componentID string, ofs *overlayfs.OverlayFs) *valueobject.SourceFilesystem {
		dirs := b.theBigFs.OverlayDirs[componentID]
		log.Printf("create view for %s with dirs: %v", componentID, dirs)
		return newSourceFilesystem(componentID, afero.NewBasePathFs(ofs, componentID), b.sourceFs, dirs)
	}

	b.result.Layouts = createView(module.ComponentFolderLayouts, b.theBigFs.OverlayMounts)
	b.result.Assets = createView(module.ComponentFolderAssets, b.theBigFs.OverlayMounts)
	b.result.ResourcesCache = b.theBigFs.OverlayResources
	b.result.Content = createView(module.ComponentFolderContent, b.theBigFs.OverlayMountsContent)

	return b.result, nil
}

func (b *sourceFilesystemsBuilder) createMainOverlayFs() (*valueobject.FilesystemsCollector, error) {
	collector := &valueobject.FilesystemsCollector{
		SourceProject:        b.sourceFs,
		OverlayMounts:        overlayfs.New(overlayfs.Options{}),
		OverlayMountsContent: overlayfs.New(overlayfs.Options{DirsMerger: valueobject.LanguageDirsMerger}),
		OverlayResources:     overlayfs.New(overlayfs.Options{FirstWritable: true}),

		OverlayDirs: make(map[string][]valueobject.FileMetaInfo),
	}
	err := b.createOverlayFs(collector)

	return collector, err
}

func (b *sourceFilesystemsBuilder) createOverlayFs(collector *valueobject.FilesystemsCollector) error {
	for _, md := range b.modules.All() {

		var (
			fromTo        []valueobject.RootMapping
			fromToContent []valueobject.RootMapping
		)

		absPathify := func(path string) string {
			if filepath.IsAbs(path) {
				return path
			}
			return paths.AbsPathify(md.Dir(), path)
		}

		log.Println("=== module: ", md)
		for _, mount := range md.Mounts() {
			log.Println("--- module mount --> ", "source", mount.Source(), "target", mount.Target(), "lang", mount.Lang())
			rm := valueobject.RootMapping{
				From: mount.Target(),             // content
				To:   absPathify(mount.Source()), // mycontent
				Meta: &valueobject.FileMeta{
					Classifier: valueobject.ContentClassContent,
				},
			}

			isContentMount := b.isContentMount(mount.Target())
			if isContentMount {
				fromToContent = append(fromToContent, rm)
			} else if b.isStaticMount(mount.Target()) {
				continue
			} else {
				fromTo = append(fromTo, rm)
			}
		}

		// Module Fs is not find in collector
		rmfs := newRootMappingFs(collector.SourceProject, fromTo...)
		rmfsContent := newRootMappingFs(collector.SourceProject, fromToContent...)

		collector.AddDirs(rmfs)        // add other folders, /layouts etc
		collector.AddDirs(rmfsContent) // only has /content, why need to go through all components?

		collector.OverlayMounts = collector.OverlayMounts.Append(rmfs)
		collector.OverlayMountsContent = collector.OverlayMountsContent.Append(rmfsContent)
		collector.OverlayResources = collector.OverlayResources.Append(
			valueobject.NewBasePathFs(collector.SourceProject, absPathify(module.FolderResources)))
	}

	return nil
}

func (b *sourceFilesystemsBuilder) isContentMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderContent)
}

func (b *sourceFilesystemsBuilder) isStaticMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderStatic)
}
