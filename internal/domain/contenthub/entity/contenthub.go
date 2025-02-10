package entity

import (
	"context"
	"errors"
	"fmt"
	"github.com/bep/logg"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	goTmpl "html/template"
	"time"
)

type ContentHub struct {
	Fs contenthub.FsService

	// ExecTemplate handling.
	TemplateExecutor contenthub.Template

	*Cache
	*Translator

	*Search

	*PageMap
	*PageFinder

	*Title

	Log      loggers.Logger `json:"-"`
	pagesLog logg.LevelLogger
}

func (ch *ContentHub) RenderString(ctx context.Context, args ...any) (goTmpl.HTML, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", errors.New("RenderString want 1 or 2 arguments")
	}

	sidx := 1
	if len(args) == 1 {
		sidx = 0
	} else {
		_, ok := args[0].(map[string]any)
		if !ok {
			return "", errors.New("first argument must be a map")
		}

		return "", errors.New("RenderString not implemented yet")
	}

	contentToRenderv := args[sidx]
	contentStr := contentToRenderv.(string)

	fmi := ch.Fs.NewFileMetaInfoWithContent(contentStr)
	file, err := valueobject.NewFileInfo(fmi)
	if err != nil {
		return "", err
	}

	ps, err := ch.Cache.GetOrCreateResource(helpers.MD5String(contentStr), func() (contenthub.PageSource, error) {
		return newPageSource(file, ch.Cache)
	})
	if err != nil {
		return "", err
	}

	p, err := ch.PageBuilder.WithSource(ps.(*Source)).Build()
	if err != nil {
		return "", err
	}

	os, err := p.PageOutputs()
	if err != nil {
		return "", err
	}

	for _, o := range os {
		c, err := o.Content()
		if err != nil {
			return "", err
		}
		return c.(goTmpl.HTML), nil
	}

	return "", nil
}

func (ch *ContentHub) ProcessPages(exec contenthub.Template) error {
	ch.pagesLog = ch.Log.InfoCommand("ContentHub.ProcessPages")
	defer loggers.TimeTrackf(ch.pagesLog, time.Now(), nil, "")

	ch.TemplateExecutor = exec
	ch.PageMap.PageBuilder.TemplateSvc = exec

	if err := ch.process(); err != nil {
		return fmt.Errorf("process: %w", err)
	}

	return nil
}

func (ch *ContentHub) CollectPages(exec contenthub.Template) error {
	ch.pagesLog = ch.Log.InfoCommand("ContentHub.CollectPages")
	defer loggers.TimeTrackf(ch.pagesLog, time.Now(), nil, "")

	ch.TemplateExecutor = exec
	ch.PageMap.PageBuilder.TemplateSvc = exec

	if err := ch.process(); err != nil {
		return fmt.Errorf("process: %w", err)
	}

	if err := ch.assemble(); err != nil {
		return fmt.Errorf("assemble: %w", err)
	}

	return nil
}

func (ch *ContentHub) process() error {
	processLog := ch.pagesLog.WithField("step", "process")
	defer loggers.TimeTrackf(processLog, time.Now(), nil, "")

	c := &pagesCollector{
		m:  ch.PageMap,
		fs: ch.Fs,

		ctx:        context.Background(),
		infoLogger: ch.pagesLog,
	}

	if err := c.Collect(); err != nil {
		return err
	}

	return nil
}

func (ch *ContentHub) assemble() error {
	processLog := ch.pagesLog.WithField("step", "assemble")
	defer loggers.TimeTrackf(processLog, time.Now(), nil, "")

	if err := ch.PageMap.Assemble(); err != nil {
		return err
	}
	return nil
}

func (ch *ContentHub) GetPageSources(page contenthub.Page) ([]contenthub.PageSource, error) {
	keyPage := page.Paths().Base()
	if keyPage == "/" {
		keyPage = ""
	}
	key := keyPage + "/get-sources-for-page"
	v, err := ch.Cache.GetOrCreateResources(key, func() ([]contenthub.PageSource, error) {
		return ch.PageMap.getResourcesForPage(page)
	})
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (ch *ContentHub) GlobalPages(langIndex int) contenthub.Pages {
	return ch.PageMap.getPagesInSection(
		langIndex,
		pageMapQueryPagesInSection{
			pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
				Path:    "",
				KeyPart: "global",
				Include: pagePredicates.ShouldListGlobal,
			},
			Recursive:   true,
			IncludeSelf: true,
		},
	)
}

func (ch *ContentHub) GlobalRegularPages() contenthub.Pages {
	return ch.PageMap.getPagesInSection(
		0,
		pageMapQueryPagesInSection{
			pageMapQueryPagesBelowPath: pageMapQueryPagesBelowPath{
				Path:    "",
				KeyPart: "global",
				Include: pagePredicates.ShouldListGlobal.And(pagePredicates.KindPage),
			},
			Recursive:   true,
			IncludeSelf: true,
		},
	)
}
