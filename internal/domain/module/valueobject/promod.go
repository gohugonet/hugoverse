package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/module"

type ProjectModule struct {
	*Module
}

func (pm *ProjectModule) ApplyComponentsMounts(components []module.Component) {
	for _, component := range components {
		if component.Dir() == "" {
			continue
		}
		pm.MountDirs = append(pm.MountDirs, Mount{
			SourcePath: component.Dir(),
			TargetPath: component.Name(),
			Language:   component.Language(),
		})
	}
}
