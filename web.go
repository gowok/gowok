package gowok

import (
	"context"
	"encoding/json"
	"fmt"
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
		ctx := &WebCtx{r.Context(), &HttpResponse{ngamux.Res(w)}, &HttpRequest{ngamux.Req(r)}}
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

type HttpResponse struct {
	*ngamux.Response
}

type HttpRequest struct {
	*ngamux.Request
}

type WebCtx struct {
	ctx context.Context
	res *HttpResponse
	req *HttpRequest
}

func (ctx WebCtx) Req() *HttpRequest {
	return ctx.req
}

func (ctx WebCtx) Res() *HttpResponse {
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

func (ctx HttpResponse) Ok(body ...any) error {
	HttpOk(ctx, body...)
	return nil
}

func (ctx HttpResponse) BadRequest(body any) error {
	HttpBadRequest(ctx, body)
	return nil
}

func (ctx HttpResponse) Unauthorized() error {
	HttpUnauthorized(ctx)
	return nil
}

func (ctx HttpResponse) NotFound() error {
	HttpNotFound(ctx)
	return nil
}

func (ctx HttpResponse) InternalServerError(body any) error {
	HttpNotFound(ctx)
	return nil
}

func (ctx HttpResponse) Created(body any) error {
	HttpNotFound(ctx)
	return nil
}

func (ctx HttpResponse) Forbidden() error {
	HttpNotFound(ctx)
	return nil
}

func (ctx HttpResponse) Conflict(rw http.ResponseWriter, body any) error {
	HttpNotFound(ctx)
	return nil
}

func (ctx HttpResponse) NoContent(rw http.ResponseWriter) error {
	HttpNoContent(ctx)
	return nil
}

func (ctx HttpResponse) Accepted(rw http.ResponseWriter, body any) error {
	HttpAccepted(ctx, body)
	return nil
}

type WebSseCtx struct {
	*WebCtx
	flusher *http.Flusher
}

func (ctx *WebSseCtx) Flush() {
	(*ctx.flusher).Flush()
}

func (ctx *WebSseCtx) Publish(message []byte) error {
	fmt.Fprintf(ctx.res, "data: %s\n\n", string(message))
	(*ctx.flusher).Flush()
	return nil
}

func (ctx *WebSseCtx) Emit(event string, message []byte) error {
	fmt.Fprintf(ctx.res, "event: %s\ndata: %s\n\n", event, string(message))
	(*ctx.flusher).Flush()
	return nil
}

func (ctx *WebSseCtx) PublishRaw(format string, a ...any) error {
	fmt.Fprintf(ctx.res, format, a...)
	(*ctx.flusher).Flush()
	return nil
}

func HandlerSse(handler func(ctx *WebSseCtx)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			HttpInternalServerError(w, errors.ErrStreamingUnsupported)
			return
		}

		handler(&WebSseCtx{
			WebCtx: &WebCtx{
				res: &HttpResponse{ngamux.Res(w)},
				req: &HttpRequest{ngamux.Req(r)},
				ctx: r.Context(),
			},
			flusher: &flusher,
		})
	}
}
