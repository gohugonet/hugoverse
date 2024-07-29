package entity

import (
	"context"
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	goTmpl "html/template"
	"time"
)

type ContentHub struct {
	Fs contenthub.FsService

	// ExecTemplate handling.
	TemplateExecutor contenthub.Template

	*PageMap

	*Title

	Log      loggers.Logger `json:"-"`
	pagesLog logg.LevelLogger
}

func (ch *ContentHub) RenderString(ctx context.Context, args ...any) (goTmpl.HTML, error) {
	//TODO
	return "", nil
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
