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
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	xmaps "golang.org/x/exp/maps"
	"path/filepath"
	"reflect"
	"sort"
)

const noConfigFileErrInfo = "\"Unable to locate config file or config directory. \\n "

type ConfigLoader struct {
	Cfg config.Provider

	config.SourceDescriptor

	BaseDirs valueobject.BaseDirs

	Logger loggers.Logger
}

func (cl *ConfigLoader) loadConfigMain() error {
	filename := cl.SourceDescriptor.Filename()
	if err := cl.loadConfig(filename); err != nil {
		fmt.Println(err)
		return err
	}
	if err := cl.applyDefaultConfig(); err != nil {
		return err
	}
	cl.Cfg.SetDefaultMergeStrategy()

	if !cl.Cfg.IsSet("languages") {
		// We need at least one
		lang := cl.Cfg.GetString("defaultContentLanguage")
		cl.Cfg.Set("languages", maps.Params{lang: maps.Params{}})
	}

	return nil
}

func (cl *ConfigLoader) deleteMergeStrategies() {
	cl.Cfg.WalkParams(func(params ...maps.KeyParams) bool {
		params[len(params)-1].Params.DeleteMergeStrategy()
		return false
	})
}

func (cl *ConfigLoader) loadConfig(configName string) error {
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
	fmt.Println(configName, filename, cl.SourceDescriptor.Fs())

	if filename == "" {
		return errors.New(noConfigFileErrInfo)
	}

	m, err := cl.loadConfigFromFile(cl.SourceDescriptor.Fs(), filename)
	if err != nil {
		return err
	}

	// Set overwrites keys of the same name, recursively.
	cl.Cfg.Set("", m)

	return nil
}

func (cl *ConfigLoader) loadConfigFromFile(fs afero.Fs, filename string) (map[string]any, error) {
	m, err := metadecoders.Default.UnmarshalFileToMap(fs, filename)
	if err != nil {
		return nil, err
	}
	return m, nil
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

func (cl *ConfigLoader) loadConfigAggregator() (*entity.Config, error) {
	all := &valueobject.BaseConfig{}
	if err := cl.decodeConfig(cl.Cfg, all, nil); err != nil {
		return nil, err
	}

	if err := all.CompileConfig(cl.Logger); err != nil {
		return nil, err
	}

	langConfigMap := make(map[string]*valueobject.BaseConfig)
	languagesConfig := cl.Cfg.GetStringMap("languages")

	cfg := cl.Cfg
	for k, v := range languagesConfig {
		mergedConfig := NewDefaultProvider()
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

							fmt.Println(456, kk, mergedConfigEntry)
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
				langConfigMap[k] = all
				continue
			}

			// Create a copy of the complete config and replace the root keys with the language specific ones.
			clone := all.CloneForLang()

			if err := cl.decodeConfig(mergedConfig, clone, differentRootKeys); err != nil {
				return nil, fmt.Errorf("failed to decode config for language %q: %w", k, err)
			}
			if err := clone.CompileConfig(cl.Logger); err != nil {
				return nil, err
			}

			langConfigMap[k] = clone

		case maps.ParamsMergeStrategy:
		default:
			panic(fmt.Sprintf("unknown type in languages config: %T", v))

		}
	}
	cm := &entity.Config{
		Base:              all,
		LanguageConfigMap: langConfigMap,
	}

	fmt.Printf("base: %+v\n", cm.Base)
	fmt.Printf("LanguageConfigMap en: %+v\n", cm.LanguageConfigMap["en"])

	return cm, nil
}

func (cl *ConfigLoader) decodeConfig(p config.Provider, target *valueobject.BaseConfig, keys []string) error {
	var decoderSetups []valueobject.DecodeWeight

	valueobject.LowerDecoderKey()
	if len(keys) == 0 {
		for _, v := range valueobject.AllDecoderSetups {
			decoderSetups = append(decoderSetups, v)
		}
	} else {
		for _, key := range keys {
			if v, found := valueobject.AllDecoderSetups[key]; found {
				decoderSetups = append(decoderSetups, v)
			} else {
				cl.Logger.Warnf("Skip unknown config key %q", key)
			}
		}
	}

	// Sort them to get the dependency order right.
	sort.Slice(decoderSetups, func(i, j int) bool {
		ki, kj := decoderSetups[i], decoderSetups[j]
		if ki.Weight == kj.Weight {
			return ki.Key < kj.Key
		}
		return ki.Weight < kj.Weight
	})

	for _, v := range decoderSetups {
		p := valueobject.DecodeConfig{
			Provider:   p,
			BaseConfig: target,
			Fs:         cl.SourceDescriptor.Fs(),
		}
		if err := v.Decode(v, p); err != nil {
			return fmt.Errorf("failed to decode %q: %w", v.Key, err)
		}
	}

	return nil
}
