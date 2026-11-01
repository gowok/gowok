package logger

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gowok/fp/slices"
)

func Configure(handlers ...slog.Handler) func() {
	return func() {
		slog.SetDefault(slog.New(newHandler(handlers...)))
	}
}

type handler struct {
	mut      *sync.Mutex
	handlers []slog.Handler
}

func newHandler(handlers ...slog.Handler) *handler {
	return &handler{mut: &sync.Mutex{}, handlers: handlers}
}

func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	return slices.Any(h.handlers, func(hh slog.Handler, _ int) bool {
		return hh.Enabled(ctx, l)
	})
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	h.mut.Lock()
	defer h.mut.Unlock()

	for _, hh := range h.handlers {
		err := hh.Handle(ctx, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := slices.MapIndex(h.handlers, func(hh slog.Handler, _ int) slog.Handler {
		return hh.WithAttrs(attrs)
	})

	return &handler{mut: h.mut, handlers: handlers}
}

func (h *handler) WithGroup(group string) slog.Handler {
	handlers := slices.MapIndex(h.handlers, func(hh slog.Handler, _ int) slog.Handler {
		return hh.WithGroup(group)
	})

	return &handler{mut: h.mut, handlers: handlers}
}
