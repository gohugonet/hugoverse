package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"strings"
)

type DecodeConfig struct {
	Provider   config.Provider
	BaseConfig *BaseConfig
	Fs         afero.Fs
}

type DecodeWeight struct {
	Key         string
	Decode      func(DecodeWeight, DecodeConfig) error
	GetCompiler func(c *BaseConfig) config.Compiler
	Weight      int
}

var AllDecoderSetups = map[string]DecodeWeight{
	"": {
		Key:    "",
		Weight: -100, // Always first.
		Decode: func(d DecodeWeight, p DecodeConfig) error {
			if err := mapstructure.WeakDecode(p.Provider.Get(""), &p.BaseConfig.RootConfig); err != nil {
				return err
			}

			// This need to match with Lang which is always lower case.
			p.BaseConfig.RootConfig.DefaultContentLanguage = strings.ToLower(p.BaseConfig.RootConfig.DefaultContentLanguage)

			return nil
		},
	},
	"module": {
		Key: "module",
		Decode: func(d DecodeWeight, p DecodeConfig) error {
			var err error
			p.BaseConfig.Module, err = DecodeModuleConfig(p.Provider)
			return err
		},
	},
	"languages": {
		Key: "languages",
		Decode: func(d DecodeWeight, p DecodeConfig) error {
			var err error
			m := p.Provider.GetStringMap(d.Key)
			if len(m) == 1 {
				// In v0.112.4 we moved this to the language config, but it's very commmon for mono language sites to have this at the top level.
				var first maps.Params
				var ok bool
				for _, v := range m {
					first, ok = v.(maps.Params)
					if ok {
						break
					}
				}
				if first != nil {
					if _, found := first["languagecode"]; !found {
						code := p.Provider.GetString("languagecode")
						if code == "" {
							code = "en-US"
						}
						first["languagecode"] = code
					}
				}
			}
			p.BaseConfig.Languages, err = DecodeLanguageConfig(m)
			if err != nil {
				return err
			}

			// Validate defaultContentLanguage.
			var found bool
			for lang := range p.BaseConfig.Languages {
				if lang == p.BaseConfig.DefaultContentLanguage {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("config value %q for defaultContentLanguage does not match any language definition", p.BaseConfig.DefaultContentLanguage)
			}

			return nil
		},
	},
}

func LowerDecoderKey() {
	for k, v := range AllDecoderSetups {
		// Verify that k and v.key is all lower case.
		if k != strings.ToLower(k) {
			panic(fmt.Sprintf("key %q is not lower case", k))
		}
		if v.Key != strings.ToLower(v.Key) {
			panic(fmt.Sprintf("key %q is not lower case", v.Key))
		}

		if k != v.Key {
			panic(fmt.Sprintf("key %q is not the same as the map key %q", k, v.Key))
		}
	}
}
