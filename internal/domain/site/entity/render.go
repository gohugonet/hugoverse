package entity

import (
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/pkg/env"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"sync"
)

type Render struct {
	numWorkers int

	pages chan *Page

	results chan error
	errs    chan error

	wg *sync.WaitGroup

	log logg.LevelLogger
}

func newRender(log logg.LevelLogger) *Render {
	numWorkers := env.GetNumWorkerMultiplier()

	return &Render{
		numWorkers: numWorkers,
		pages:      make(chan *Page, numWorkers),

		results: make(chan error),
		errs:    make(chan error),

		log: log,
	}
}

func (r *Render) close() {
	close(r.pages)
}

func (r *Render) startRenderPages() {
	go r.errorCollator(r.results, r.errs)

	r.wg = &sync.WaitGroup{}

	for i := 0; i < r.numWorkers; i++ {
		r.wg.Add(1)
		go r.renderPage()
	}

	r.wg.Wait()

	close(r.results)
}

func (r *Render) renderPage() {
	defer r.wg.Done()

	for p := range r.pages {
		if err := p.render(); err != nil {
			r.results <- err
		}
	}
}

func (r *Render) errorCollator(results <-chan error, errs chan<- error) {
	var errors []error
	for e := range results {
		errors = append(errors, e)
	}

	errs <- r.pickOneAndLogTheRest(errors)

	close(errs)
}

func (r *Render) pickOneAndLogTheRest(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var i int

	for j, err := range errors {
		// If this is in server mode, we want to return an error to the client
		// with a file context, if possible.
		if herrors.UnwrapFileError(err) != nil {
			i = j
			break
		}
	}

	// Log the rest, but add a threshold to avoid flooding the log.
	const errLogThreshold = 5

	for j, err := range errors {
		if j == i || err == nil {
			continue
		}

		if j >= errLogThreshold {
			break
		}

		r.log.WithError(err)
	}

	return errors[i]
}
