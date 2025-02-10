package valueobject

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdfriday/hugoverse/pkg/collections"
	"github.com/mdfriday/hugoverse/pkg/hexec"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"io"
	"os"
	"os/exec"
	"strings"
)

type GoClient struct {
	Exec *hexec.Exec
	Dir  string
	// Environment variables used in "go get" etc.
	Environ []string

	// Set if we get a exec.ErrNotFound when running Go, which is most likely
	// due to being run on a system without Go installed. We record it here
	// so we can give an instructional error at the end if module/theme
	// resolution fails.
	goBinaryStatus goBinaryStatus

	Logger loggers.Logger
}

func (c *GoClient) listGoMods() (GoModules, error) {
	downloadModules := func(modules ...string) error {
		args := []string{"mod", "download", "-modcacherw"}
		args = append(args, modules...)
		out := io.Discard
		err := c.runGo(context.Background(), out, args...)
		if err != nil {
			return fmt.Errorf("failed to download modules: %w", err)
		}
		return nil
	}

	if err := downloadModules(); err != nil {
		return nil, err
	}

	listAndDecodeModules := func(handle func(m *GoModule) error, modules ...string) error {
		b := &bytes.Buffer{}
		args := []string{"list", "-m", "-json"}
		if len(modules) > 0 {
			args = append(args, modules...)
		} else {
			args = append(args, "all")
		}
		err := c.runGo(context.Background(), b, args...)
		if err != nil {
			return fmt.Errorf("failed to list modules: %w", err)
		}

		dec := json.NewDecoder(b)
		for {
			m := &GoModule{}
			if err := dec.Decode(m); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to decode modules list: %w", err)
			}

			if err := handle(m); err != nil {
				return err
			}
		}
		return nil
	}

	var modules GoModules
	err := listAndDecodeModules(func(m *GoModule) error {
		modules = append(modules, m)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// From Go 1.17, go lazy loads transitive dependencies.
	// That does not work for us.
	// So, download these modules and update the Dir in the modules list.
	var modulesToDownload []string
	for _, m := range modules {
		if m.Dir == "" {
			modulesToDownload = append(modulesToDownload, fmt.Sprintf("%s@%s", m.Path, m.Version))
		}
	}

	c.Logger.Println("Module to download: ", modulesToDownload)
	if len(modulesToDownload) > 0 {
		if err := downloadModules(modulesToDownload...); err != nil {
			return nil, err
		}
		err := listAndDecodeModules(func(m *GoModule) error {
			if mm := modules.GetByPath(m.Path); mm != nil {
				mm.Dir = m.Dir
			}
			return nil
		}, modulesToDownload...)
		if err != nil {
			return nil, err
		}
	}

	return modules, err
}

func (c *GoClient) runGo(ctx context.Context, stdout io.Writer, args ...string) error {
	if c.goBinaryStatus != 0 {
		return nil
	}

	stderr := new(bytes.Buffer)

	argsv := collections.StringSliceToInterfaceSlice(args)
	argsv = append(argsv, hexec.WithEnviron(c.Environ))
	argsv = append(argsv, hexec.WithStderr(io.MultiWriter(stderr, os.Stderr)))
	argsv = append(argsv, hexec.WithStdout(stdout))
	argsv = append(argsv, hexec.WithDir(c.Dir))
	argsv = append(argsv, hexec.WithContext(ctx))

	cmd, err := c.Exec.New("go", argsv...)
	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		var ee *exec.Error
		if errors.As(err, &ee) && errors.Is(ee.Err, exec.ErrNotFound) {
			c.goBinaryStatus = goBinaryStatusNotFound
			return nil
		}

		if strings.Contains(stderr.String(), "invalid version: unknown revision") {
			// See https://github.com/gohugoio/hugo/issues/6825
			c.Logger.Println(errInvalidInfo)
		}

		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		if !ok {
			return fmt.Errorf("failed to execute 'go %v': %s %T", args, err, err)
		}

		// Too old Go version
		if strings.Contains(stderr.String(), "flag provided but not defined") {
			c.goBinaryStatus = goBinaryStatusTooOld
			return nil
		}

		return fmt.Errorf("go command failed: %s", stderr)

	}

	return nil
}

func (c *GoClient) Get(args ...string) error {
	var hasD bool
	for _, arg := range args {
		if arg == "-d" {
			hasD = true
			break
		}
	}
	if !hasD {
		// go get without the -d flag does not make sense to us, as
		// it will try to build and install go packages.
		args = append([]string{"-d"}, args...)
	}
	if err := c.runGo(context.Background(), c.Logger.Out(), append([]string{"get"}, args...)...); err != nil {
		return fmt.Errorf("failed to get %q: %w", args, err)
	}
	return nil
}

func (c *GoClient) WrapModuleNotFound(err error) error {
	err = fmt.Errorf(err.Error()+": %w", ErrNotExist)
	baseMsg := "we found a go.mod file in your project, but"

	switch c.goBinaryStatus {
	case goBinaryStatusNotFound:
		return fmt.Errorf(baseMsg+" you need to install Go to use it. See https://golang.org/dl/ : %q", err)
	case goBinaryStatusTooOld:
		return fmt.Errorf(baseMsg+" you need to a newer version of Go to use it. See https://golang.org/dl/ : %w", err)
	default:
		return err
	}
}
