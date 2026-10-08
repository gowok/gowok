package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gowok/gowok/errors"
	"github.com/gowok/gowok/web/request"
	"github.com/gowok/gowok/web/response"
)

type Ctx struct {
	ctx context.Context
	res *response.Response
	req *request.Request
}

func NewCtx(ctx context.Context, w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{
		r.Context(),
		response.New(w),
		request.New(r),
	}
}

func (ctx Ctx) Req() *request.Request {
	return ctx.req
}

func (ctx Ctx) Res() *response.Response {
	return ctx.res
}

func (ctx *Ctx) Write(content []byte) (int, error) {
	return ctx.res.Write(content)
}

func (ctx *Ctx) SetContext(ctxNew context.Context) {
	ctx.ctx = ctxNew
}

func (ctx Ctx) Deadline() (time.Time, bool) {
	return ctx.ctx.Deadline()
}

func (ctx Ctx) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx Ctx) Err() error {
	return ctx.ctx.Err()
}

func (ctx Ctx) Value(key any) any {
	return ctx.ctx.Value(key)
}

func (ctx *Ctx) SetValue(key, value any) {
	ctx.SetContext(context.WithValue(ctx.ctx, key, value))
}

type CtxSse struct {
	*Ctx
	flusher *http.Flusher
}

func NewCtxSse(ctx *Ctx) (*CtxSse, error) {
	flusher, ok := ctx.res.ResponseWriter.(http.Flusher)
	if !ok {
		return nil, errors.ErrStreamingUnsupported
	}

	return &CtxSse{
		Ctx:     ctx,
		flusher: &flusher,
	}, nil
}

func (ctx *CtxSse) Flush() {
	(*ctx.flusher).Flush()
}

func (ctx *CtxSse) Publish(message []byte) error {
	return ctx.PublishRaw("data: %s\n\n", string(message))
}

func (ctx *CtxSse) Emit(event string, message []byte) error {
	return ctx.PublishRaw("event: %s\ndata: %s\n\n", event, string(message))
}

func (ctx *CtxSse) PublishRaw(format string, a ...any) error {
	_, _ = fmt.Fprintf(ctx.res, format, a...)
	(*ctx.flusher).Flush()
	return nil
}
