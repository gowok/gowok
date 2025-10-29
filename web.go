package gowok

import (
	"net/http"

	"github.com/gowok/gowok/errors"
	"github.com/gowok/gowok/web"
	"github.com/ngamux/ngamux"
)

func Router() *ngamux.HttpServeMux {
	Get()
	return web.Router()
}

func Server() *http.Server {
	Get()
	return web.Server()
}

func Handler(handler func(ctx *web.Ctx) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := web.NewCtx(r.Context(), w, r)
		err := handler(ctx)
		if err != nil {
			switch e := err.(type) {
			case errors.Error:
				ngamux.Res(w).JSON(e)
			default:
				HttpInternalServerError(w, err)
			}
		}
	}
}

func HandlerSse(handler func(ctx *web.CtxSse)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx, err := web.NewCtxSse(web.NewCtx(r.Context(), w, r))
		if err != nil {
			HttpInternalServerError(w, err)
			return
		}

		handler(ctx)
	}
}
