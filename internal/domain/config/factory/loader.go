package factory

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/entity"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	xmaps "golang.org/x/exp/maps"
	"path/filepath"
	"reflect"
)

const noConfigFileErrInfo = "\"Unable to locate config file or config directory. \\n "

type ConfigLoader struct {
	Cfg config.Provider

	config.SourceDescriptor

	BaseDirs valueobject.BaseDirs

	Logger loggers.Logger
}

func (cl *ConfigLoader) loadConfigByDefault() (config.Provider, error) {
	filename := cl.SourceDescriptor.Filename()
	if err := cl.loadProvider(filename); err != nil {
		return nil, err
	}
	if err := cl.applyDefaultConfig(); err != nil {
		return nil, err
	}
	cl.Cfg.SetDefaultMergeStrategy()

	if !cl.Cfg.IsSet("languages") {
		// We need at least one
		lang := cl.Cfg.GetString("defaultContentLanguage")
		cl.Cfg.Set("languages", maps.Params{lang: maps.Params{}})
	}

	return cl.Cfg, nil
}

func (cl *ConfigLoader) deleteMergeStrategies() {
	cl.Cfg.WalkParams(func(params ...maps.KeyParams) bool {
		params[len(params)-1].Params.DeleteMergeStrategy()
		return false
	})
}

func (cl *ConfigLoader) loadProvider(configName string) error {
	baseDir := cl.BaseDirs.WorkingDir
	var baseFilename string
	if filepath.IsAbs(configName) {
		baseFilename = configName
	} else {
		baseFilename = filepath.Join(baseDir, configName)
	}
	var filename string
	if paths.ExtNoDelimiter(configName) != "" {
		exists, _ := afero.Exists(cl.SourceDescriptor.Fs(), baseFilename)
		if exists {
			filename = baseFilename
		}
	}

	if filename == "" {
		return errors.New(noConfigFileErrInfo)
	}

	m, err := valueobject.LoadConfigFromFile(cl.SourceDescriptor.Fs(), filename)
	if err != nil {
		return err
	}

	// Set overwrites keys of the same name, recursively.
	cl.Cfg.Set("", m)

	return nil
}

func (cl *ConfigLoader) applyDefaultConfig() error {
	defaultSettings := maps.Params{
		"baseURL":                              "",
		"cleanDestinationDir":                  false,
		"watch":                                false,
		"contentDir":                           "content",
		"resourceDir":                          "resources",
		"publishDir":                           "public",
		"publishDirOrig":                       "public",
		"themesDir":                            "themes",
		"assetDir":                             "assets",
		"layoutDir":                            "layouts",
		"i18nDir":                              "i18n",
		"dataDir":                              "data",
		"archetypeDir":                         "archetypes",
		"configDir":                            "config",
		"staticDir":                            "static",
		"buildDrafts":                          false,
		"buildFuture":                          false,
		"buildExpired":                         false,
		"params":                               maps.Params{},
		"environment":                          config.EnvironmentProduction,
		"uglyURLs":                             false,
		"verbose":                              false,
		"ignoreCache":                          false,
		"canonifyURLs":                         false,
		"relativeURLs":                         false,
		"removePathAccents":                    false,
		"titleCaseStyle":                       "AP",
		"taxonomies":                           maps.Params{"tag": "tags", "category": "categories"},
		"permalinks":                           maps.Params{},
		"sitemap":                              maps.Params{"priority": -1, "filename": "sitemap.xml"},
		"menus":                                maps.Params{},
		"disableLiveReload":                    false,
		"pluralizeListTitles":                  true,
		"capitalizeListTitles":                 true,
		"forceSyncStatic":                      false,
		"footnoteAnchorPrefix":                 "",
		"footnoteReturnLinkContents":           "",
		"newContentEditor":                     "",
		"paginate":                             10,
		"paginatePath":                         "page",
		"summaryLength":                        70,
		"rssLimit":                             -1,
		"sectionPagesMenu":                     "",
		"disablePathToLower":                   false,
		"hasCJKLanguage":                       false,
		"enableEmoji":                          false,
		"defaultContentLanguage":               "en",
		"defaultContentLanguageInSubdir":       false,
		"enableMissingTranslationPlaceholders": false,
		"enableGitInfo":                        false,
		"ignoreFiles":                          make([]string, 0),
		"disableAliases":                       false,
		"debug":                                false,
		"disableFastRender":                    false,
		"timeout":                              "30s",
		"timeZone":                             "",
		"enableInlineShortcodes":               false,
	}

	cl.Cfg.SetDefaults(defaultSettings)

	return nil
}

func (cl *ConfigLoader) assembleConfig(c *entity.Config) error {
	c.Language.Default = cl.Cfg.GetString("defaultContentLanguage")

	return cl.decodeConfig(cl.Cfg, c)
}

func (cl *ConfigLoader) decodeConfig(p config.Provider, target *entity.Config) error {
	r, err := valueobject.DecodeRoot(p)
	if err != nil {
		return err
	}
	target.Root.RootConfig = r
	target.Root.RootConfig.BaseDirs = cl.BaseDirs

	cs, err := valueobject.DecodeCachesConfig(cl.SourceDescriptor.Fs(), p, cl.BaseDirs)
	if err != nil {
		return err
	}
	target.Caches.CachesConfig = cs

	sec, err := valueobject.DecodeSecurityConfig(p)
	if err != nil {
		return err
	}
	target.Security.SecurityConfig = sec

	img, err := valueobject.DecodeImagingConfig(p)
	if err != nil {
		return err
	}
	target.Imaging.ImagingConfigInternal = img

	m, err := valueobject.DecodeModuleConfig(p)
	if err != nil {
		return err
	}
	target.Module.ModuleConfig = m

	languages, err := valueobject.DecodeLanguageConfig(p)
	if err != nil {
		return err
	}
	target.Language.Configs = languages
	if err := target.Validate(); err != nil {
		return err
	}

	langConfigMap := make(map[string]valueobject.RootConfig)
	languagesConfig := cl.Cfg.GetStringMap("languages")

	cfg := cl.Cfg
	for k, v := range languagesConfig {
		mergedConfig := valueobject.NewDefaultProvider()
		var differentRootKeys []string
		switch x := v.(type) {
		case maps.Params:
			var params maps.Params
			pv, found := x["params"]
			if found {
				params = pv.(maps.Params)
			} else {
				params = maps.Params{
					maps.MergeStrategyKey: maps.ParamsMergeStrategyDeep,
				}
				x["params"] = params
			}

			for kk, vv := range x {
				if kk == "_merge" {
					continue
				}

				mergedConfig.Set(kk, vv)
				rootv := cfg.Get(kk)
				if rootv != nil && cfg.IsSet(kk) {
					// This overrides a root key and potentially needs a merge.
					if !reflect.DeepEqual(rootv, vv) {
						switch vvv := vv.(type) {
						case maps.Params:
							differentRootKeys = append(differentRootKeys, kk)

							// Use the language value as base.
							mergedConfigEntry := xmaps.Clone(vvv)
							// Merge in the root value.
							maps.MergeParams(mergedConfigEntry, rootv.(maps.Params))

							mergedConfig.Set(kk, mergedConfigEntry)
						default:
							// Apply new values to the root.
							differentRootKeys = append(differentRootKeys, "")
						}
					}
				} else {
					switch vv.(type) {
					case maps.Params:
						differentRootKeys = append(differentRootKeys, kk)
					default:
						// Apply new values to the root.
						differentRootKeys = append(differentRootKeys, "")
					}
				}
			}
			differentRootKeys = helpers.UniqueStringsSorted(differentRootKeys)

			if len(differentRootKeys) == 0 {
				langConfigMap[k] = target.RootConfig
				continue
			}

			clone, err := valueobject.DecodeRoot(mergedConfig)
			cl.Logger.Printf("hugoverse: merging config for language `%s`, %+v", k, clone)
			if err != nil {
				return err
			}

			langConfigMap[k] = clone

		case maps.ParamsMergeStrategy:
		default:
			panic(fmt.Sprintf("unknown type in languages config: %T", v))

		}
	}
	target.Language.RootConfigs = langConfigMap

	return nil
}
