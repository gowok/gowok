package web

import (
	"net/http"

	"github.com/gowok/gowok/errors"
)

func Handler(handler func(ctx *Ctx) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewCtx(r.Context(), w, r)
		err := handler(ctx)
		if err != nil {
			switch e := err.(type) {
			case errors.Error:
				ctx.Res().JSON(e)
			default:
				_ = ctx.Res().InternalServerError(err)
			}
		}
	}
}

func HandlerSSE(handler func(ctx *CtxSse)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx, err := NewCtxSse(NewCtx(r.Context(), w, r))
		if err != nil {
			_ = ctx.Res().InternalServerError(err)
			return
		}

		handler(ctx)
	}
}
