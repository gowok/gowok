package logger

import (
	"context"
	"log/slog"
	"sync"
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
	for _, hh := range h.handlers {
		if hh.Enabled(ctx, l) {
			return true
		}
	}

	return false
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
	for i, hh := range h.handlers {
		h.handlers[i] = hh.WithAttrs(attrs)
	}

	return h
}

func (h *handler) WithGroup(group string) slog.Handler {
	for i, hh := range h.handlers {
		h.handlers[i] = hh.WithGroup(group)
	}

	return h
}
