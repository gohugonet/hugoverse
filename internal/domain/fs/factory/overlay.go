package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
)

func CreateMainOverlayFs(ofs *entity.OriginFs, mods module.Modules) (*valueobject.FilesystemsCollector, error) {
	collector := &valueobject.FilesystemsCollector{
		SourceProject: ofs.Source,
		SourceModules: ofs.Source,

		OverlayMounts:        overlayfs.New(overlayfs.Options{}),
		OverlayMountsContent: overlayfs.New(overlayfs.Options{DirsMerger: valueobject.LanguageDirsMerger}),
		OverlayMountsStatic:  overlayfs.New(overlayfs.Options{DirsMerger: valueobject.LanguageDirsMerger}),
		OverlayFull:          overlayfs.New(overlayfs.Options{}),
		OverlayResources:     overlayfs.New(overlayfs.Options{FirstWritable: true}),

		AbsResourcesDir: ofs.AbsResourcesDir,
	}

	if err := collector.Collect(mods); err != nil {
		return nil, err
	}

	return collector, nil
}
