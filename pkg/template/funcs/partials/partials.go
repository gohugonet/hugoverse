// Copyright 2017 The Hugo Authors. All rights reserved.
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

// Package partials provides template functions for working with reusable
// templates.
package partials

import (
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"strings"
	"time"

	"github.com/bep/lazycache"
)

type partialCacheKey struct {
	Name     string
	Variants []any
}
type includeResult struct {
	name     string
	result   any
	mangager identity.Manager
	err      error
}

func (k partialCacheKey) Key() string {
	if k.Variants == nil {
		return k.Name
	}
	return identity.HashString(append([]any{k.Name}, k.Variants...)...)
}

func (k partialCacheKey) templateName() string {
	if !strings.HasPrefix(k.Name, "partials/") {
		return "partials/" + k.Name
	}
	return k.Name
}

// partialCache represents a LRU cache of partials.
type partialCache struct {
	cache *lazycache.Cache[string, includeResult]
}

func (p *partialCache) clear() {
	p.cache.DeleteFunc(func(s string, r includeResult) bool {
		return true
	})
}

// New returns a new instance of the templates-namespaced template functions.
func New(cb func(ctx context.Context, name string, data any) (tmpl, res string, err error)) *Namespace {
	// This lazycache was introduced in Hugo 0.111.0.
	// We're going to expand and consolidate all memory caches in Hugo using this,
	// so just set a high limit for now.
	lru := lazycache.New(lazycache.Options[string, includeResult]{MaxEntries: 1000})

	cache := &partialCache{cache: lru}
	defer cache.clear()

	return &Namespace{
		cachedPartials:   cache,
		templateExecutor: cb,
	}
}

type TemplateExecutor func(ctx context.Context, name string, data any) (tmpl, res string, err error)

// Namespace provides template functions for the "templates" namespace.
type Namespace struct {
	cachedPartials   *partialCache
	templateExecutor TemplateExecutor
}

// contextWrapper makes room for a return value in a partial invocation.
type contextWrapper struct {
	Arg    any
	Result any
}

// Set sets the return value and returns an empty string.
func (c *contextWrapper) Set(in any) string {
	c.Result = in
	return ""
}

// Include executes the named partial.
// If the partial contains a return statement, that value will be returned.
// Else, the rendered output will be returned:
// A string if the partial is a text/template, or template.HTML when html/template.
// Note that ctx is provided by Hugo, not the end user.
func (ns *Namespace) Include(ctx context.Context, name string, contextList ...any) (any, error) {
	res := ns.includWithTimeout(ctx, name, contextList...)
	if res.err != nil {
		return nil, res.err
	}

	//TODO track metrics for result

	return res.result, nil
}

func (ns *Namespace) includWithTimeout(ctx context.Context, name string, dataList ...any) includeResult {
	// Create a new context with a timeout not connected to the incoming context.
	// TODO hardcode timeout here for now
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res := make(chan includeResult, 1)

	go func() {
		res <- ns.include(ctx, name, dataList...)
	}()

	select {
	case r := <-res:
		return r
	case <-timeoutCtx.Done():
		err := timeoutCtx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("partial %q timed out after %s. This is most likely due to infinite recursion. If this is just a slow template, you can try to increase the 'timeout' config setting", name, "30s")
		}
		return includeResult{err: err}
	}
}

// include is a helper function that lookups and executes the named partial.
// Returns the final template name and the rendered output.
func (ns *Namespace) include(ctx context.Context, name string, dataList ...any) includeResult {
	var data any
	if len(dataList) > 0 {
		data = dataList[0]
	}

	var n string
	if strings.HasPrefix(name, "partials/") {
		n = name
	} else {
		n = "partials/" + name
	}

	tmpl, res, err := ns.templateExecutor(ctx, n, data)
	if err != nil {
		return includeResult{err: err}
	}

	return includeResult{
		name:   tmpl,
		result: res,
	}
}
