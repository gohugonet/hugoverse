package entity

import (
	"fmt"
	"github.com/bep/godartsass/v2"
	"github.com/bep/logg"
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/internal/domain/resources/valueobject"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/media"
	"io"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// See https://github.com/sass/dart-sass-embedded/issues/24
// Note: This prefix must be all lower case.
const dartSassStdinPrefix = "hugostdin:"

type SassClient struct {
	BinaryFound bool
	AllowedExec bool

	FsService resources.Fs

	// The Dart Client requires a os/exec process, so  only
	// create it if we really need it.
	// This is mostly to avoid creating one per site build test.
	once sync.Once
	// One of these are non-nil.
	transpiler *godartsass.Transpiler
}

func (c *SassClient) Open() error {
	if c.AllowedExec && c.BinaryFound {
		var (
			transpiler *godartsass.Transpiler
			err        error
			infol      = loggers.NewDefault().InfoCommand("Dart Sass")
			warnl      = loggers.NewDefault().WarnCommand("Dart Sass")
		)

		c.once.Do(func() {
			if valueobject.IsDartSassV2() {
				transpiler, err = godartsass.Start(godartsass.Options{
					DartSassEmbeddedFilename: valueobject.DartSassBinaryName,
					LogEventHandler: func(event godartsass.LogEvent) {
						message := strings.ReplaceAll(event.Message, dartSassStdinPrefix, "")
						switch event.Type {
						case godartsass.LogEventTypeDebug:
							// Log as Info for now, we may adjust this if it gets too chatty.
							infol.Log(logg.String(message))
						default:
							// The rest are either deprecations or @warn statements.
							warnl.Log(logg.String(message))
						}
					},
				})
				c.transpiler = transpiler
			} else {
				panic(fmt.Sprintf("Unexpected Dart Sass version, only supported v2: %s", valueobject.DartSassBinaryName))
			}
		})

		if err != nil {
			return fmt.Errorf("dart sass client error: %s, %s", err.Error(), c.generateErrorMessage(c.BinaryFound, c.AllowedExec))
		}
	}

	return nil
}

func (c *SassClient) generateErrorMessage(binaryFound, allowedExec bool) string {
	if binaryFound && allowedExec {
		return "Binary found and execution allowed."
	} else if binaryFound && !allowedExec {
		return "Binary found but execution not allowed."
	} else {
		return "Binary not found."
	}
}

func (c *SassClient) Close() error {
	if c.transpiler != nil {
		return c.transpiler.Close()
	}
	return nil
}

func (c *SassClient) ToCSS(res resources.Resource, args map[string]any) (resources.Resource, error) {
	transRes := res.(Transformer)
	return transRes.Transform(&scssTransformation{c: c, optsm: args, log: loggers.NewDefault()})
}

func (c *SassClient) toCSS(args godartsass.Args, src io.Reader) (godartsass.Result, error) {
	in := helpers.ReaderToString(src)

	args.Source = in

	var (
		err error
		res godartsass.Result
	)

	res, err = c.transpiler.Execute(args)

	if err != nil {
		if err.Error() == "unexpected EOF" {
			//lint:ignore ST1005 end user message.
			return res, fmt.Errorf("got unexpected EOF when executing %q. The user running hugo must have read and execute permissions on this program. With execute permissions only, this error is thrown", valueobject.DartSassBinaryName)
		}
		return res, herrors.NewFileErrorFromFileInErr(err, c.FsService.AssetsFs(), herrors.OffsetMatcher)
	}

	return res, err
}

type scssTransformation struct {
	optsm map[string]any
	c     *SassClient

	log loggers.Logger
}

const transformationName = "tocss-dart"

func (t *scssTransformation) Key() valueobject.ResourceTransformationKey {
	return valueobject.NewResourceTransformationKey(transformationName, t.optsm)
}

func (t *scssTransformation) Transform(ctx *valueobject.ResourceTransformationCtx) error {
	ctx.Target.OutMediaType = media.Builtin.CSSType

	opts, err := valueobject.DecodeDartSassOptions(t.optsm)
	if err != nil {
		return err
	}

	if opts.TargetPath != "" {
		ctx.Target.OutPath = opts.TargetPath
	} else {
		ctx.ReplaceOutPathExtension(".css")
	}

	baseDir := path.Dir(ctx.SourcePath())
	filename := dartSassStdinPrefix

	if ctx.SourcePath() != "" {
		filename += t.c.FsService.AssetsFsRealFilename(ctx.SourcePath())
	}

	args := godartsass.Args{
		URL:          filename,
		IncludePaths: t.c.FsService.AssetsFsRealDirs(baseDir),
		ImportResolver: valueobject.ImportResolver{
			BaseDir:           baseDir,
			FsService:         t.c.FsService,
			DependencyManager: ctx.DepSvc.DependencyManager(),

			VarsStylesheet: godartsass.Import{Content: valueobject.CreateVarsStyleSheet(opts.Vars)},
		},
		OutputStyle:             godartsass.ParseOutputStyle(opts.OutputStyle),
		EnableSourceMap:         opts.EnableSourceMap,
		SourceMapIncludeSources: opts.SourceMapIncludeSources,
	}

	// Append any workDir relative include paths
	for _, ip := range opts.IncludePaths {
		info, err := t.c.FsService.AssetsFs().Stat(filepath.Clean(ip))
		if err == nil {
			filename := info.(fs.FileMetaInfo).FileName()
			args.IncludePaths = append(args.IncludePaths, filename)
		}
	}

	if ctx.Source.InMediaType.SubType == media.Builtin.SASSType.SubType {
		args.SourceSyntax = godartsass.SourceSyntaxSASS
	}

	res, err := t.c.toCSS(args, ctx.Source.From)
	if err != nil {
		t.log.Printf("toCSS error: %s", err)
		return err
	}

	out := res.CSS

	_, err = io.WriteString(ctx.Target.To, out)
	if err != nil {
		t.log.Errorf("toCSS write string error: %s", err)
		return err
	}

	if opts.EnableSourceMap && res.SourceMap != "" {
		target := ctx.Target.OutPath + ".map"

		if err := ctx.PubSvc.PublishContentToTarget(res.SourceMap, target); err != nil {
			return err
		}
		_, err = fmt.Fprintf(ctx.Target.To, "\n\n/*# sourceMappingURL=%s */", path.Base(ctx.Target.OutPath)+".map")
	}

	return err
}
