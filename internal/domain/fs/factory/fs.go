package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
)

func New(dir fs.Dir, mods module.Modules) (*entity.Fs, error) {
	f := &entity.Fs{
		OriginFs: NewOriginFs(dir),
		Service:  &entity.Service{},
	}

	collector, err := CreateMainOverlayFs(f.OriginFs, mods)
	if err != nil {
		return nil, err
	}
	// create component fs

	f.Archetypes = newComponentFs(files.ComponentFolderArchetypes, collector.OverlayMounts)
	f.Layouts = newComponentFs(files.ComponentFolderLayouts, collector.OverlayMounts)
	f.Assets = newComponentFs(files.ComponentFolderAssets, collector.OverlayMounts)
	f.ResourcesCache = collector.OverlayResources
	f.RootFss = collector.RootFss

	// data and i18n  needs a different merge strategy.
	overlayMountsPreserveDupes := collector.OverlayMounts.WithDirsMerger(valueobject.AppendDirsMerger)
	f.Data = newComponentFs(files.ComponentFolderData, overlayMountsPreserveDupes)
	f.I18n = newComponentFs(files.ComponentFolderI18n, overlayMountsPreserveDupes)
	f.AssetsWithDuplicatesPreserved = newComponentFs(files.ComponentFolderAssets, overlayMountsPreserveDupes)

	f.Content = newComponentFs(files.ComponentFolderContent, collector.OverlayMountsContent)
	f.Work = valueobject.NewReadOnlyFs(collector.OverlayFull)

	// Create static filesystem(s)
	ms := make(map[string]*valueobject.ComponentFs)
	ms[""] = newComponentFs(files.ComponentFolderStatic, collector.OverlayMountsStatic)
	f.Static = ms

	return f, nil
}

func newComponentFs(component string, overlayFs *overlayfs.OverlayFs) *valueobject.ComponentFs {
	return valueobject.NewComponentFs(component, overlayFs)
}
