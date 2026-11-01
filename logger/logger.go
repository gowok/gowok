package logger

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gowok/fp/slices"
)

func Configure(handlers ...slog.Handler) func() {
	return func() {
		slog.SetDefault(slog.New(&handler{handlers: handlers}))
	}
}

type handler struct {
	mux      sync.Mutex
	handlers []slog.Handler
}

func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	return slices.Any(h.handlers, func(hh slog.Handler, _ int) bool {
		return hh.Enabled(ctx, l)
	})
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	h.mux.Lock()
	defer h.mux.Unlock()

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

	return &handler{handlers: handlers}
}

func (h *handler) WithGroup(group string) slog.Handler {
	handlers := slices.MapIndex(h.handlers, func(hh slog.Handler, _ int) slog.Handler {
		return hh.WithGroup(group)
	})

	return &handler{handlers: handlers}
}
