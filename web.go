package gowok

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

func Handler(handler func(ctx *WebCtx) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &WebCtx{r.Context(), ngamux.Res(w), ngamux.Req(r)}
		err := handler(ctx)
		if err != nil {
			if gowokErr, ok := err.(errors.Error); ok {
				err = json.NewEncoder(ctx.res).Encode(gowokErr)
				if err == nil {
					return
				}
			}
			HttpInternalServerError(w, err)
		}
	}
}

type WebCtx struct {
	ctx context.Context
	res *ngamux.Response
	req *ngamux.Request
}

func (ctx WebCtx) Req() *ngamux.Request {
	return ctx.req
}

func (ctx WebCtx) Res() *ngamux.Response {
	return ctx.res
}

func (ctx *WebCtx) Write(content []byte) (int, error) {
	return ctx.res.Write(content)
}

func (ctx *WebCtx) SetContext(ctxNew context.Context) {
	ctx.ctx = ctxNew
}

func (ctx WebCtx) Deadline() (time.Time, bool) {
	return ctx.ctx.Deadline()
}

func (ctx WebCtx) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx WebCtx) Err() error {
	return ctx.ctx.Err()
}

func (ctx WebCtx) Value(key any) any {
	return ctx.ctx.Value(key)
}

func (ctx *WebCtx) SetValue(key, value any) {
	ctx.SetContext(context.WithValue(ctx.ctx, key, value))
}

func (ctx WebCtx) Ok(body ...any) error {
	HttpOk(ctx.res, body...)
	return nil
}

func (ctx WebCtx) HttpBadRequest(body any) error {
	HttpBadRequest(ctx.res, body)
	return nil
}

func (ctx WebCtx) HttpUnauthorized() error {
	HttpUnauthorized(ctx.res)
	return nil
}

func (ctx WebCtx) HttpNotFound() error {
	HttpNotFound(ctx.res)
	return nil
}

func (ctx WebCtx) HttpInternalServerError(body any) error {
	HttpNotFound(ctx.res)
	return nil
}

func (ctx WebCtx) HttpCreated(body any) error {
	HttpNotFound(ctx.res)
	return nil
}

func (ctx WebCtx) HttpForbidden() error {
	HttpNotFound(ctx.res)
	return nil
}

func (ctx WebCtx) HttpConflict(rw http.ResponseWriter, body any) error {
	HttpNotFound(ctx.res)
	return nil
}

func (ctx WebCtx) NoContent(rw http.ResponseWriter) error {
	HttpNoContent(ctx.res)
	return nil
}

func (ctx WebCtx) Accepted(rw http.ResponseWriter, body any) error {
	HttpAccepted(ctx.res, body)
	return nil
}
