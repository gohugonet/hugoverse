package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/content/valueobject"
	"github.com/mdfriday/hugoverse/pkg/env"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"sync"
)

type Writer struct {
	numWorkers int

	files chan *valueobject.File

	results chan error
	errs    chan error

	wg *sync.WaitGroup

	log loggers.Logger
}

func newWriter(log loggers.Logger) *Writer {
	numWorkers := env.GetNumWorkerMultiplier()

	return &Writer{
		numWorkers: numWorkers,
		files:      make(chan *valueobject.File, numWorkers),

		results: make(chan error),
		errs:    make(chan error),

		log: log,
	}
}

func (r *Writer) close() {
	close(r.files)
}

func (r *Writer) startDumpFiles() {
	go r.errorCollator(r.results, r.errs)

	r.wg = &sync.WaitGroup{}

	for i := 0; i < r.numWorkers; i++ {
		r.wg.Add(1)
		go r.dumpFiles()
	}

	r.wg.Wait()

	close(r.results)
}

func (r *Writer) dumpFiles() {
	defer r.wg.Done()

	for p := range r.files {
		if err := p.Dump(); err != nil {
			r.results <- err
		}
	}
}

func (r *Writer) errorCollator(results <-chan error, errs chan<- error) {
	var errors []error
	for e := range results {
		errors = append(errors, e)
	}

	errs <- r.pickOneAndLogTheRest(errors)

	close(errs)
}

func (r *Writer) pickOneAndLogTheRest(errors []error) error {
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

		r.log.Errorln(err)
	}

	return errors[i]
}
