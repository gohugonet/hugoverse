package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/resource/resourcehelpers"
	"github.com/spf13/cast"
	"reflect"
)

// New returns a new instance of the resources-namespaced template functions.
func New(resource Resource) (*Namespace, error) {
	return &Namespace{
		resourceService: resource,
	}, nil
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	resourceService Resource
}

// Get locates the filename given in Hugo's assets filesystem
// and creates a Resource object that can be used for further transformations.
func (ns *Namespace) Get(filename any) resources.Resource {
	filenamestr, err := cast.ToStringE(filename)
	if err != nil {
		panic(err)
	}

	r, err := ns.resourceService.GetResource(filenamestr)
	if err != nil {
		panic(err)
	}

	return r
}

// Copy copies r to the new targetPath in s.
func (ns *Namespace) Copy(s any, r resources.Resource) (resources.Resource, error) {
	targetPath, err := cast.ToStringE(s)
	if err != nil {
		panic(err)
	}
	return ns.resourceService.Copy(r, targetPath)
}

// GetMatch finds the first Resource matching the given pattern, or nil if none found.
//
// It looks for files in the assets file system.
//
// See Match for a more complete explanation about the rules used.
func (ns *Namespace) GetMatch(pattern any) resources.Resource {
	patternStr, err := cast.ToStringE(pattern)
	if err != nil {
		panic(err)
	}

	r, err := ns.resourceService.GetMatch(patternStr)
	if err != nil {
		panic(err)
	}

	return r
}

// Minify minifies the given Resource using the MediaType to pick the correct
// minifier.
func (ns *Namespace) Minify(r resources.Resource) (resources.Resource, error) {
	return ns.resourceService.Minify(r)
}

// ExecuteAsTemplate creates a Resource from a Go template, parsed and executed with
// the given data, and published to the relative target path.
func (ns *Namespace) ExecuteAsTemplate(ctx context.Context, args ...any) (resources.Resource, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("must provide targetPath, the template data context and a Resource object")
	}
	targetPath, err := cast.ToStringE(args[0])
	if err != nil {
		return nil, err
	}
	data := args[1]

	r, ok := args[2].(resources.Resource)
	if !ok {
		return nil, fmt.Errorf("type %T not supported in Resource transformations", args[2])
	}

	return ns.resourceService.ExecuteAsTemplate(ctx, r, targetPath, data)
}

// Fingerprint transforms the given Resource with a MD5 hash of the content in
// the RelPermalink and Permalink.
func (ns *Namespace) Fingerprint(args ...any) (resources.Resource, error) {
	if len(args) < 1 {
		return nil, errors.New("must provide a Resource object")
	}

	if len(args) > 2 {
		return nil, errors.New("must not provide more arguments than Resource and hash algorithm")
	}

	var algo string
	resIdx := 0

	if len(args) == 2 {
		resIdx = 1
		var err error
		algo, err = cast.ToStringE(args[0])
		if err != nil {
			return nil, err
		}
	}

	r, ok := args[resIdx].(resources.Resource)
	if !ok {
		return nil, fmt.Errorf("%T can not be transformed", args[resIdx])
	}

	return ns.resourceService.Fingerprint(r, algo)
}

func (ns *Namespace) Sass(args ...any) (resources.Resource, error) {
	return ns.ToCSS(args...)
}

// ToCSS converts the given Resource to CSS. You can optional provide an Options object
// as second argument. As an option, you can e.g. specify e.g. the target path (string)
// for the converted CSS resource.
func (ns *Namespace) ToCSS(args ...any) (resources.Resource, error) {
	if len(args) > 2 {
		return nil, errors.New("must not provide more arguments than resource object and options")
	}

	const (
		// Transpiler implementation can be controlled from the client by
		// setting the 'transpiler' option.
		// Default is currently 'libsass', but that may change.
		transpilerDart    = "dartsass"
		transpilerLibSass = "libsass" // not supported anymore
	)

	var (
		r          resources.Resource
		m          map[string]any
		targetPath string
		err        error
		ok         bool
	)

	r, targetPath, ok = resourcehelpers.ResolveIfFirstArgIsString(args)

	if !ok {
		r, m, err = resourcehelpers.ResolveArgs(args)
		if err != nil {
			return nil, err
		}
	}

	if m != nil {
		if t, found := maps.LookupEqualFold(m, "transpiler"); found {
			switch t {
			case transpilerDart:
				log := loggers.NewDefault()
				log.Debugf("using transpiler %q", cast.ToString(t))
			default:
				return nil, fmt.Errorf("unsupported transpiler %q; valid values are %q", t, transpilerDart)
			}
		}
	}

	if m == nil {
		m = make(map[string]any)
	}
	if targetPath != "" {
		m["targetPath"] = targetPath
	}

	return ns.resourceService.ToCSS(r, m)
}

// Concat concatenates a slice of Resource objects. These resources must
// (currently) be of the same Media Type.
func (ns *Namespace) Concat(targetPathIn any, r any) (resources.Resource, error) {
	targetPath, err := cast.ToStringE(targetPathIn)
	if err != nil {
		return nil, err
	}

	rv := reflect.ValueOf(r)
	if rv.Kind() != reflect.Slice {
		return nil, errors.New("expected slice of Resource objects, received " + rv.Kind().String() + " instead")
	}

	var rr []resources.Resource
	for i := 0; i < rv.Len(); i++ {
		rr = append(rr, rv.Index(i).Interface().(resources.Resource))
	}

	if len(rr) == 0 {
		return nil, errors.New("must provide one or more Resource objects to concat")
	}

	return ns.resourceService.Concat(targetPath, rr)
}
