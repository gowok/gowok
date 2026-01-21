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
				if e.Code() != 0 {
					ctx.Res().Status(e.Code())
				}
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

		c := NewCtx(r.Context(), w, r)
		ctx, err := NewCtxSse(c)
		if err != nil {
			_ = c.Res().InternalServerError(err)
			return
		}

		handler(ctx)
	}
}
