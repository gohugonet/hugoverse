// Copyright 2024 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package entity

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/media"
	"io"
	"path"
	"path/filepath"
	"regexp"
)

type JsClient struct {
	c *valueobject.BuildClient
}

func NewJsClient(fs resources.Fs, log loggers.Logger) *JsClient {
	return &JsClient{
		c: valueobject.NewBuildClient(fs, log),
	}
}

func (c *JsClient) ProcessJs(res resources.Resource, opts map[string]any) (resources.Resource, error) {
	transRes := res.(Transformer)
	return transRes.Transform(
		&buildTransformation{c: c, optsm: opts},
	)
}

func (c *JsClient) transform(opts valueobject.Options, transformCtx *valueobject.ResourceTransformationCtx) (api.BuildResult, error) {
	if transformCtx.DepSvc.DependencyManager() != nil {
		opts.DependencyManager = transformCtx.DepSvc.DependencyManager()
	}

	opts.StdinSourcePath = transformCtx.SourcePath()

	result, err := c.c.Build(opts)
	if err != nil {
		return result, err
	}

	if opts.ExternalOptions.SourceMap == "linked" || opts.ExternalOptions.SourceMap == "external" {
		content := string(result.OutputFiles[1].Contents)
		if opts.ExternalOptions.SourceMap == "linked" {
			symPath := path.Base(transformCtx.Target.OutPath) + ".map"
			re := regexp.MustCompile(`//# sourceMappingURL=.*\n?`)
			content = re.ReplaceAllString(content, "//# sourceMappingURL="+symPath+"\n")
		}

		target := transformCtx.Target.OutPath + ".map"
		if err = transformCtx.PubSvc.PublishContentToTarget(string(result.OutputFiles[0].Contents), target); err != nil {
			return result, err
		}
		_, err := transformCtx.Target.To.Write([]byte(content))
		if err != nil {
			return result, err
		}
	} else {
		_, err := transformCtx.Target.To.Write(result.OutputFiles[0].Contents)
		if err != nil {
			return result, err
		}

	}
	return result, nil
}

type buildTransformation struct {
	optsm map[string]any
	c     *JsClient
}

func (t *buildTransformation) Key() valueobject.ResourceTransformationKey {
	return valueobject.NewResourceTransformationKey("jsbuild", t.optsm)
}

func (t *buildTransformation) Transform(ctx *valueobject.ResourceTransformationCtx) error {
	ctx.Target.OutMediaType = media.Builtin.JavascriptType

	var opts valueobject.Options

	if t.optsm != nil {
		optsExt, err := valueobject.DecodeExternalOptions(t.optsm)
		if err != nil {
			return err
		}
		opts.ExternalOptions = optsExt
	}

	if opts.TargetPath != "" {
		ctx.Target.OutPath = opts.TargetPath
	} else {
		ctx.ReplaceOutPathExtension(".js")
	}

	src, err := io.ReadAll(ctx.Source.From)
	if err != nil {
		return err
	}

	opts.SourceDir = filepath.FromSlash(path.Dir(ctx.SourcePath()))
	opts.Contents = string(src)
	opts.MediaType = ctx.Source.InMediaType
	opts.Stdin = true

	_, err = t.c.transform(opts, ctx)

	return err
}
