// Package multi implements a handler which invokes a number of handlers.
package multi

import (
	"github.com/bep/logg"
)

// Handler implementation.
type Handler struct {
	Handlers []logg.Handler
}

// New handler.
func New(h ...logg.Handler) *Handler {
	return &Handler{
		Handlers: h,
	}
}

// HandleLog implements logg.Handler.
func (h *Handler) HandleLog(e *logg.Entry) error {
	for _, handler := range h.Handlers {
		// TODO(tj): maybe just write to stderr here, definitely not ideal
		// to miss out logging to a more critical handler if something
		// goes wrong
		if err := handler.HandleLog(e); err != nil {
			return err
		}
	}

	return nil
}
