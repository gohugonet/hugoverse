package factory

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/spf13/afero"
	"strings"
)

func newSourceFilesystem(name string, fs afero.Fs, dirs []valueobject.FileMetaInfo) *valueobject.SourceFilesystem {
	sf := &valueobject.SourceFilesystem{
		Name: name,
		Fs:   fs,
		Dirs: dirs,
	}

	info, _, err := valueobject.LstatIfPossible(fs, "")
	if err != nil {
		fmt.Println("0000111", info, err)
	}

	return sf
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
		return newSourceFilesystem(componentID, afero.NewBasePathFs(ofs, componentID), dirs)
	}

	b.result.Layouts = createView(module.ComponentFolderLayouts, b.theBigFs.OverlayMounts)
	b.result.Content = createView(module.ComponentFolderContent, b.theBigFs.OverlayMountsContent)

	return b.result, nil
}

func (b *sourceFilesystemsBuilder) createMainOverlayFs() (*valueobject.FilesystemsCollector, error) {
	collector := &valueobject.FilesystemsCollector{
		SourceProject:        b.sourceFs,
		OverlayMounts:        overlayfs.New([]afero.Fs{}),
		OverlayMountsContent: overlayfs.New([]afero.Fs{}),
		OverlayDirs:          make(map[string][]valueobject.FileMetaInfo),
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

		for _, mount := range md.Mounts() {
			rm := valueobject.RootMapping{
				From: mount.Target, // content
				To:   mount.Source, // mycontent
				Meta: &valueobject.FileMeta{
					Classifier: valueobject.ContentClassContent,
				},
			}

			isContentMount := b.isContentMount(mount.Target)
			if isContentMount {
				fromToContent = append(fromToContent, rm)
			} else if b.isStaticMount(mount.Target) {
				continue
			} else {
				fromTo = append(fromTo, rm)
			}
		}

		rmfs := newRootMappingFs(collector.SourceProject, fromTo...)
		rmfsContent := newRootMappingFs(collector.SourceProject, fromToContent...)

		collector.AddDirs(rmfs)        // add other folders, /layouts etc
		collector.AddDirs(rmfsContent) // only has /content, why need to go through all components?

		collector.OverlayMounts = collector.OverlayMounts.Append(rmfs)
		collector.OverlayMountsContent = collector.OverlayMountsContent.Append(rmfsContent)
	}

	return nil
}

func (b *sourceFilesystemsBuilder) isContentMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderContent)
}

func (b *sourceFilesystemsBuilder) isStaticMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderStatic)
}
