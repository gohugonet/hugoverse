package entity

import (
	"context"
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
	goTmpl "html/template"
	"time"
)

type ContentHub struct {
	Fs contenthub.FsService

	// ExecTemplate handling.
	TemplateExecutor contenthub.Template

	*Cache

	*PageMap

	*Title

	Log      loggers.Logger `json:"-"`
	pagesLog logg.LevelLogger
}

func (ch *ContentHub) RenderString(ctx context.Context, args ...any) (goTmpl.HTML, error) {
	//TODO
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
	keyPage := page.Paths().Path()
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

func (ch *ContentHub) GetPageFromPath(path string) (contenthub.Page, error) {
	p := paths.Parse(files.ComponentFolderContent, path)
	n := ch.PageMap.TreePages.Get(p.Base()) // TODO, shape?

	if n != nil {
		ps, found := n.getPage()
		if !found {
			return valueobject.NilPage, nil
		}

		return ps, nil
	}

	return nil, nil
}

func (ch *ContentHub) GlobalPages() contenthub.Pages {
	return ch.PageMap.getPagesInSection(
		0,
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
