package valueobject

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/cache"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	FilecacheRootDirname = "filecache"
)

const (
	resourcesGenDir = ":resourceDir/_gen"
	cacheDirProject = ":cacheDir/:project"
)

var defaultCacheConfig = FileCacheConfig{
	MaxAge: -1, // Never expire
	Dir:    cacheDirProject,
}

var defaultCacheConfigs = CachesConfig{
	cache.KeyModules: {
		MaxAge: -1,
		Dir:    ":cacheDir/modules",
	},
	cache.KeyGetJSON: defaultCacheConfig,
	cache.KeyGetCSV:  defaultCacheConfig,
	cache.KeyImages: {
		MaxAge: -1,
		Dir:    resourcesGenDir,
	},
	cache.KeyAssets: {
		MaxAge: -1,
		Dir:    resourcesGenDir,
	},
	cache.KeyGetResource: FileCacheConfig{
		MaxAge: -1, // Never expire
		Dir:    cacheDirProject,
	},
}

type CachesConfig map[string]FileCacheConfig

func (c CachesConfig) CacheDirModules() string {
	return c[cache.KeyModules].DirCompiled
}

type FileCacheConfig struct {
	// Max age of cache entries in this cache. Any items older than this will
	// be removed and not returned from the cache.
	// A negative value means forever, 0 means cache is disabled.
	// Hugo is lenient with what types it accepts here, but we recommend using
	// a duration string, a sequence of  decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	MaxAge time.Duration

	// The directory where files are stored.
	Dir         string
	DirCompiled string `json:"-"`

	// Will resources/_gen will get its own composite filesystem that
	// also checks any theme.
	IsResourceDir bool `json:"-"`
}

func DecodeCachesConfig(fs afero.Fs, p config.Provider, bcfg BaseDirs) (CachesConfig, error) {
	m := p.GetStringMap("caches")

	c := make(CachesConfig)
	valid := make(map[string]bool)
	// Add defaults
	for k, v := range defaultCacheConfigs {
		c[k] = v
		valid[k] = true
	}

	_, isOsFs := fs.(*afero.OsFs)

	for k, v := range m {
		if _, ok := v.(maps.Params); !ok {
			continue
		}
		cc := defaultCacheConfig

		dc := &mapstructure.DecoderConfig{
			Result:           &cc,
			DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
			WeaklyTypedInput: true,
		}

		decoder, err := mapstructure.NewDecoder(dc)
		if err != nil {
			return c, err
		}

		if err := decoder.Decode(v); err != nil {
			return nil, fmt.Errorf("failed to decode filecache config: %w", err)
		}

		if cc.Dir == "" {
			return c, errors.New("must provide cache Dir")
		}

		name := strings.ToLower(k)
		if !valid[name] {
			return nil, fmt.Errorf("%q is not a valid cache name", name)
		}

		c[name] = cc
	}

	for k, v := range c {
		dir := filepath.ToSlash(filepath.Clean(v.Dir))
		hadSlash := strings.HasPrefix(dir, "/")
		parts := strings.Split(dir, "/")

		for i, part := range parts {
			if strings.HasPrefix(part, ":") {
				resolved, isResource, err := resolveDirPlaceholder(bcfg, part)
				if err != nil {
					return c, err
				}
				if isResource {
					v.IsResourceDir = true
				}
				parts[i] = resolved
			}
		}

		dir = path.Join(parts...)
		if hadSlash {
			dir = "/" + dir
		}
		v.DirCompiled = filepath.Clean(filepath.FromSlash(dir))

		if !v.IsResourceDir {
			if isOsFs && !filepath.IsAbs(v.DirCompiled) {
				return c, fmt.Errorf("%q must resolve to an absolute directory", v.DirCompiled)
			}

			// Avoid cache in root, e.g. / (Unix) or c:\ (Windows)
			if len(strings.TrimPrefix(v.DirCompiled, filepath.VolumeName(v.DirCompiled))) == 1 {
				return c, fmt.Errorf("%q is a root folder and not allowed as cache dir", v.DirCompiled)
			}
		}

		if !strings.HasPrefix(v.DirCompiled, "_gen") {
			// We do cache eviction (file removes) and since the user can set
			// his/hers own cache directory, we really want to make sure
			// we do not delete any files that do not belong to this cache.
			// We do add the cache name as the root, but this is an extra safe
			// guard. We skip the files inside /resources/_gen/ because
			// that would be breaking.
			v.DirCompiled = filepath.Join(v.DirCompiled, FilecacheRootDirname, k)
		} else {
			v.DirCompiled = filepath.Join(v.DirCompiled, k)
		}

		c[k] = v
	}

	return c, nil
}

// Resolves :resourceDir => /myproject/resources etc., :cacheDir => ...
func resolveDirPlaceholder(bcfg BaseDirs, placeholder string) (cacheDir string, isResource bool, err error) {
	switch strings.ToLower(placeholder) {
	case ":resourcedir":
		return "", true, nil
	case ":cachedir":
		return bcfg.CacheDir, false, nil
	case ":project":
		return filepath.Base(bcfg.WorkingDir), false, nil
	}

	return "", false, fmt.Errorf("%q is not a valid placeholder (valid values are :cacheDir or :resourceDir)", placeholder)
}

// GetCacheDir returns a cache dir from the given filesystem and config.
// The dir will be created if it does not exist.
func GetCacheDir(fs afero.Fs, cacheDir string) (string, error) {
	cacheDir = cacheDirDefault(cacheDir)

	if cacheDir != "" {
		exists, err := helpers.DirExists(cacheDir, fs)
		if err != nil {
			return "", err
		}
		if !exists {
			err := fs.MkdirAll(cacheDir, 0o777) // Before umask
			if err != nil {
				return "", fmt.Errorf("failed to create cache dir: %w", err)
			}
		}
		return cacheDir, nil
	}

	const hugoCacheBase = "hugo_cache"

	// Avoid filling up the home dir with Hugo cache dirs from development.
	//if !htesting.IsTest {
	userCacheDir, err := os.UserCacheDir()
	if err == nil {
		cacheDir := filepath.Join(userCacheDir, hugoCacheBase)
		if err := fs.Mkdir(cacheDir, 0o777); err == nil || os.IsExist(err) {
			return cacheDir, nil
		}
	}
	//}

	// Fall back to a cache in /tmp.
	userName := os.Getenv("USER")
	if userName != "" {
		return helpers.GetTempDir(hugoCacheBase+"_"+userName, fs), nil
	} else {
		return helpers.GetTempDir(hugoCacheBase, fs), nil
	}
}

func cacheDirDefault(cacheDir string) string {
	// Always use the cacheDir config if set.
	if len(cacheDir) > 1 {
		return helpers.AddTrailingFileSeparator(cacheDir)
	}

	// See Issue #8714.
	// Turns out that Cloudflare also sets NETLIFY=true in its build environment,
	// but all of these 3 should not give any false positives.
	if os.Getenv("NETLIFY") == "true" && os.Getenv("PULL_REQUEST") != "" && os.Getenv("DEPLOY_PRIME_URL") != "" {
		// Netlify's cache behaviour is not documented, the currently best example
		// is this project:
		// https://github.com/philhawksworth/content-shards/blob/master/gulpfile.js
		return "/opt/build/cache/hugo_cache/"
	}

	// This will fall back to an hugo_cache folder in either os.UserCacheDir or the tmp dir, which should work fine for most CI
	// providers. See this for a working CircleCI setup:
	// https://github.com/bep/hugo-sass-test/blob/6c3960a8f4b90e8938228688bc49bdcdd6b2d99e/.circleci/config.yml
	// If not, they can set the HUGO_CACHEDIR environment variable or cacheDir config key.
	return ""
}
