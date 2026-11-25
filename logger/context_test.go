package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/golang-must/must"
)

func TestContext(t *testing.T) {
	t.Run("positive/round trip", func(t *testing.T) {
		want := slog.New(slog.NewTextHandler(io.Discard, nil))

		ctx := NewContext(context.Background(), want)

		must.Equal(t, want, FromContext(ctx))
	})

	t.Run("positive/nested overrides", func(t *testing.T) {
		first := slog.New(slog.NewTextHandler(io.Discard, nil))
		second := slog.New(slog.NewTextHandler(io.Discard, nil))

		ctx := NewContext(NewContext(context.Background(), first), second)

		must.Equal(t, second, FromContext(ctx))
	})

	t.Run("positive/nil logger keeps context", func(t *testing.T) {
		ctx := context.Background()

		must.Equal(t, ctx, NewContext(ctx, nil))
	})

	t.Run("positive/missing falls back to default", func(t *testing.T) {
		must.Equal(t, slog.Default(), FromContext(context.Background()))
	})

	t.Run("positive/nil context falls back to default", func(t *testing.T) {
		must.Equal(t, slog.Default(), FromContext(context.TODO()))
	})
}
