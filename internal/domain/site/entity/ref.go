package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/text"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"path/filepath"
)

type Ref struct {
	Site       *Site
	ContentSvc site.ContentService

	NotFoundURL string
	ErrorLogger logg.LevelLogger
}

func (r *Ref) RelRefFrom(argsm map[string]any, source any) (string, error) {
	return r.relRef(argsm, source)
}

func (r *Ref) relRef(argsm map[string]any, source any) (string, error) {
	args, err := r.decodeRefArgs(argsm)
	if err != nil {
		return "", fmt.Errorf("invalid arguments to Ref: %w", err)
	}

	if args.Path == "" {
		return "", nil
	}

	// TODO, not support search byh output format yet
	return r.refLink(args.Path, source, true, args.OutputFormat)
}

func (r *Ref) decodeRefArgs(args map[string]any) (valueobject.RefArgs, error) {
	var ra valueobject.RefArgs
	err := mapstructure.WeakDecode(args, &ra)

	return ra, err
}

func (r *Ref) refLink(ref string, source any, relative bool, outputFormat string) (string, error) {
	pw, ok := source.(contenthub.PageWrapper)
	if !ok {
		return "", fmt.Errorf("source is not a PageWrapper")
	}

	p := pw.UnwrapPage()

	var refURL *url.URL

	ref = filepath.ToSlash(ref)

	refURL, err := url.Parse(ref)
	if err != nil {
		return r.NotFoundURL, err
	}

	var target contenthub.Page
	var link string

	if refURL.Path != "" {
		var err error
		target, err = r.ContentSvc.GetPageRef(p, refURL.Path, r.Site.home.Page)
		var pos text.Position
		if err != nil || target == nil {
			if po, ok := source.(text.Positioner); ok {
				pos = po.Position()
			}
		}

		if err != nil {
			r.logNotFound(refURL.Path, err.Error(), p, pos)
			return r.NotFoundURL, nil
		}

		if target == nil {
			r.logNotFound(refURL.Path, "page not found", p, pos)
			return r.NotFoundURL, nil
		}

		tsp, err := r.Site.sitePage(target)
		if err != nil {
			r.ErrorLogger.Logf("[%s] REF_NOT_FOUND: Ref %q: %s", p.PageIdentity().PageLanguage(), ref, err.Error())
			return "", err
		}

		if relative {
			link = tsp.RelPermalink()
		} else {
			link = tsp.Permalink()
		}
	}

	if refURL.Fragment != "" {
		link = link + "#" + refURL.Fragment

		// TODO, support with more detail from pageContext
	}

	return link, nil
}

func (r *Ref) logNotFound(ref, what string, p contenthub.Page, position text.Position) {
	if position.IsValid() {
		r.ErrorLogger.Logf("[%s] REF_NOT_FOUND: Ref %q: %s: %s", p.PageIdentity().PageLanguage(), ref, position.String(), what)
	} else if p == nil {
		r.ErrorLogger.Logf("[%s] REF_NOT_FOUND: Ref %q: %s", p.PageIdentity().PageLanguage(), ref, what)
	} else {
		r.ErrorLogger.Logf("[%s] REF_NOT_FOUND: Ref %q from page %q: %s", p.PageIdentity().PageLanguage(), ref, p.Path(), what)
	}
}
