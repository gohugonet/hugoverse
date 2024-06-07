package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

type FilesystemsCollector struct {
	SourceProject afero.Fs // Source for project folders
	SourceModules afero.Fs // Source for modules/themes

	OverlayMounts        *overlayfs.OverlayFs
	OverlayMountsContent *overlayfs.OverlayFs
	OverlayMountsStatic  *overlayfs.OverlayFs
	OverlayMountsFull    *overlayfs.OverlayFs
	OverlayFull          *overlayfs.OverlayFs
	OverlayResources     *overlayfs.OverlayFs

	RootFss []*RootMappingFs

	AbsResourcesDir string
}

func (c *FilesystemsCollector) Collect(mods module.Modules) error {
	for _, md := range mods.All() {
		var (
			fromTo        []RootMapping
			fromToContent []RootMapping
			fromToStatic  []RootMapping
		)

		absPathify := func(path string) (string, string) {
			if filepath.IsAbs(path) {
				return "", path
			}
			return md.Dir(), paths.AbsPathify(md.Dir(), path)
		}

		for _, mount := range md.Mounts() {
			base, absFilename := absPathify(mount.Source())

			rm := RootMapping{
				From: mount.Target(),

				To:     absFilename,
				ToBase: base,
			}

			isContentMount := isContentMount(mount.Target())

			if isContentMount {
				fromToContent = append(fromToContent, rm)
			} else if isStaticMount(mount.Target()) {
				fromToStatic = append(fromToStatic, rm)
			} else {
				fromTo = append(fromTo, rm)
			}
		}

		modBase := c.SourceProject
		sourceStatic := modBase

		rmfs, err := NewRootMappingFs(modBase, fromTo...)
		if err != nil {
			return err
		}
		rmfsContent, err := NewRootMappingFs(modBase, fromToContent...)
		if err != nil {
			return err
		}
		rmfsStatic, err := NewRootMappingFs(sourceStatic, fromToStatic...)
		if err != nil {
			return err
		}

		// We need to keep the list of directories for watching.
		c.addRootFs(rmfs)
		c.addRootFs(rmfsContent)
		c.addRootFs(rmfsStatic)

		// Do not support static per language

		getResourcesDir := func() string {
			if mods.IsProjMod(md) {
				return c.AbsResourcesDir
			}
			_, filename := absPathify(files.FolderResources)
			return filename
		}

		c.OverlayMounts = c.OverlayMounts.Append(rmfs)
		c.OverlayMountsContent = c.OverlayMountsContent.Append(rmfsContent)
		c.OverlayMountsStatic = c.OverlayMountsStatic.Append(rmfsStatic)
		c.OverlayFull = c.OverlayFull.Append(NewBasePathFs(modBase, md.Dir()))
		c.OverlayResources = c.OverlayResources.Append(NewBasePathFs(modBase, getResourcesDir()))
	}

	return nil
}

func (c *FilesystemsCollector) addRootFs(rfs *RootMappingFs) {
	c.RootFss = append(c.RootFss, rfs)
}

func isContentMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderContent)
}

func isStaticMount(target string) bool {
	return strings.HasPrefix(target, module.ComponentFolderStatic)
}
