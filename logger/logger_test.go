package logger

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang-must/must"
)

func (m *testHandler) Enabled(context.Context, slog.Level) bool { return m.enabled }
func (m *testHandler) Handle(context.Context, slog.Record) error {
	m.handled = true
	if m.err != nil {
		return m.err
	}
	return nil
}
func (m *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return m }
func (m *testHandler) WithGroup(name string) slog.Handler       { return m }
func (m *testHandler) Reset()                                   { m.handled = false }

type testHandler struct {
	enabled bool
	handled bool
	err     error
}

func TestConfigure(t *testing.T) {
	oldDefault := slog.Default()
	defer slog.SetDefault(oldDefault)

	Configure(&testHandler{})()
	must.NotEqual(t, oldDefault, slog.Default())
}

func TestHandler(t *testing.T) {
	testCases := []struct {
		name     string
		handlers []slog.Handler
		check    func(t *testing.T, h slog.Handler)
	}{
		{
			name: "positive/Enabled returns true if any handler is enabled",
			handlers: []slog.Handler{
				&testHandler{enabled: false},
				&testHandler{enabled: true},
			},
			check: func(t *testing.T, h slog.Handler) {
				must.True(t, h.Enabled(context.Background(), slog.LevelInfo))
			},
		},
		{
			name: "positive/Enabled returns false if no handler is enabled",
			handlers: []slog.Handler{
				&testHandler{enabled: false},
				&testHandler{enabled: false},
			},
			check: func(t *testing.T, h slog.Handler) {
				must.False(t, h.Enabled(context.Background(), slog.LevelInfo))
			},
		},
		{
			name: "positive/Handle calls all handlers",
			handlers: []slog.Handler{
				&testHandler{},
				&testHandler{},
			},
			check: func(t *testing.T, h slog.Handler) {
				err := h.Handle(context.Background(), slog.Record{})
				must.Nil(t, err)
				for _, hh := range h.(*handler).handlers {
					must.True(t, hh.(*testHandler).handled)
				}
			},
		},
		{
			name: "negative/Handle returns error if any handler fails",
			handlers: []slog.Handler{
				&testHandler{err: errors.New("fail")},
			},
			check: func(t *testing.T, h slog.Handler) {
				err := h.Handle(context.Background(), slog.Record{})
				must.NotNil(t, err)
				must.Equal(t, "fail", err.Error())
			},
		},
		{
			name:     "positive/WithAttrs and WithGroup return the handler",
			handlers: []slog.Handler{&testHandler{}},
			check: func(t *testing.T, h slog.Handler) {
				must.Equal(t, h, h.WithAttrs(nil))
				must.Equal(t, h, h.WithGroup("test"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &handler{handlers: tc.handlers}
			tc.check(t, h)
		})
	}
}
