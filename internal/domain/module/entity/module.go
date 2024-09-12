package entity

import (
	"fmt"
	"github.com/bep/debounce"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/internal/domain/module/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"strings"
	"time"
)

type Module struct {
	Fs            afero.Fs
	WorkingDir    string
	ModuleImports []string

	Logger loggers.Logger

	GoClient *valueobject.GoClient

	PathService module.Paths
	DirService  module.Workspace

	projMod *valueobject.ProjectModule
	modules []*valueobject.Module
	*Lang

	collector *valueobject.Collector
}

func (m *Module) Proj() module.Module {
	return m.projMod
}

func (m *Module) All() []module.Module {
	modules := []module.Module{m.projMod.Module}

	for _, mod := range m.modules {
		modules = append(modules, mod)
	}
	return modules
}

func (m *Module) IsProjMod(mod module.Module) bool {
	return m.projMod.Module == mod
}

func (m *Module) Load() error {
	defer m.Logger.PrintTimerIfDelayed(time.Now(), "hugoverse: collected modules")
	d := debounce.New(2 * time.Second)
	d(func() {
		m.Logger.Println("hugoverse: downloading modules …")
	})
	defer d(func() {})

	m.collector = valueobject.NewCollector(m.GoClient)
	if err := m.collect(); err != nil {
		return err
	}

	return nil
}

func (m *Module) collect() error {
	if err := m.collector.CollectGoModules(); err != nil {
		return err
	}
	m.projMod = &valueobject.ProjectModule{
		Module: &valueobject.Module{
			Fs:        m.Fs,
			AbsDir:    m.WorkingDir,
			GoModule:  m.collector.GetMain(),
			Parent:    nil,
			MountDirs: make([]valueobject.Mount, 0),
		}}
	if err := m.applyProjMounts(); err != nil {
		return err
	}

	if err := m.addAndRecurse(m.projMod.Module, m.ModuleImports); err != nil {
		return err
	}

	return nil
}

func (m *Module) applyProjMounts() error {
	defaultLangKey := m.DirService.DefaultLanguage()
	for _, component := range module.ComponentFolders {
		dir, err := m.DirService.GetRelDir(component, defaultLangKey)
		if err != nil {
			return err
		}
		if dir == "" {
			dir = component // No customized config, use default component name as folder name
		}

		m.projMod.AppendMount(valueobject.Mount{
			SourcePath: dir,
			TargetPath: component,
			Language:   defaultLangKey,
		})
	}

	otherLangKeys := m.DirService.OtherLanguageKeys()
	for _, l := range otherLangKeys {
		dir, err := m.DirService.GetRelDir(module.ComponentFolderContent, l)
		if err != nil {
			return err
		}
		if dir == "" {
			continue
		}

		m.projMod.AppendMount(valueobject.Mount{
			SourcePath: dir,
			TargetPath: module.ComponentFolderContent,
			Language:   l,
		})
	}

	return nil
}

func (m *Module) addAndRecurse(owner *valueobject.Module, moduleImports []string) error {
	for _, moduleImport := range moduleImports {
		if !m.collector.IsSeen(moduleImport) {
			tc, mi, err := m.add(owner, moduleImport)
			if err != nil {
				return err
			}
			if tc == nil {
				continue
			}
			if err := m.addAndRecurse(tc, mi); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) add(owner *valueobject.Module, moduleImport string) (*valueobject.Module, []string, error) {
	moduleDir, err := m.getImportModuleDir(moduleImport)
	if err != nil {
		return nil, nil, err
	}

	mod := m.collector.GetGoModule(moduleImport)
	if mod == nil {
		return nil, nil, fmt.Errorf("module %s not found", moduleImport)
	}

	mo := &valueobject.Module{
		Fs:       m.Fs,
		AbsDir:   moduleDir,
		Path:     moduleImport,
		GoModule: mod,
		Parent:   owner,
	}

	moImportPaths, err := m.PathService.GetImports(mo.AbsDir)

	if err := mo.ApplyMounts(valueobject.Import{Path: moduleImport}); err != nil {
		return nil, nil, err
	}

	m.modules = append(m.modules, mo)
	return mo, moImportPaths, nil
}

// Get runs "go get" with the supplied arguments.
func (m *Module) Get(args ...string) error {
	return m.GoClient.Get(args...)
}

func (m *Module) getImportModuleDir(modulePath string) (string, error) {
	var (
		moduleDir string
		mod       *valueobject.GoModule
	)

	if moduleDir == "" {
		var versionQuery string
		mod = m.collector.GetGoModule(modulePath)
		if mod != nil {
			moduleDir = mod.Dir
			versionQuery = mod.Version
		}

		if moduleDir == "" {
			if valueobject.IsProbablyModule(modulePath) {
				if versionQuery == "" {
					// See https://golang.org/ref/mod#version-queries
					// This will select the latest release-version (not beta etc.).
					versionQuery = "upgrade"
				}

				m.Logger.Println("hugoverse: get module from", modulePath)
				if err := m.Get(fmt.Sprintf("%s@%s", modulePath, versionQuery)); err != nil {
					return "", err
				}
				if err := m.collector.CollectGoModules(); err != nil {
					return "", err
				}

				mod = m.collector.GetGoModule(modulePath)
				if mod != nil {
					moduleDir = mod.Dir
				}
			}

			// Fall back to project/themes/<mymodule>
			if moduleDir == "" {
				return "", m.GoClient.WrapModuleNotFound(fmt.Errorf(
					`module %q not found; only support go module to load theme at this moment`, modulePath))
			}
		}
	}

	if found, _ := afero.Exists(m.Fs, moduleDir); !found {
		err := m.GoClient.WrapModuleNotFound(fmt.Errorf("%q not found", moduleDir))
		return "", err
	}

	if !strings.HasSuffix(moduleDir, fs.FilepathSeparator) {
		moduleDir += fs.FilepathSeparator
	}

	return moduleDir, nil
}
