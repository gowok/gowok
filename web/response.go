package web

import (
	"encoding/json"
	"net/http"

	"github.com/ngamux/ngamux"
)

type HttpResponse struct {
	*ngamux.Response
}

func NewResponse(w http.ResponseWriter) *HttpResponse {
	return &HttpResponse{ngamux.Res(w)}
}

func (ctx HttpResponse) bodyParse(res *ngamux.Response, body ...any) {
	var body1 any = ""
	if len(body) > 0 {
		body1 = body[0]
	}

	switch b := body1.(type) {
	case string:
		res.Text(b)
	case json.Marshaler:
		res.JSON(b)
	case error:
		res.Text(b.Error())
	default:
		res.JSON(b)
	}
}

func (ctx HttpResponse) Ok(body ...any) error {
	res := ctx.Status(http.StatusOK)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) BadRequest(body ...any) error {
	res := ctx.Status(http.StatusBadRequest)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) Unauthorized(body ...any) error {
	res := ctx.Status(http.StatusUnauthorized)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) NotFound(body ...any) error {
	res := ctx.Status(http.StatusNotFound)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) InternalServerError(body ...any) error {
	res := ctx.Status(http.StatusInternalServerError)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) Created(body ...any) error {
	res := ctx.Status(http.StatusCreated)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) Forbidden(body ...any) error {
	res := ctx.Status(http.StatusForbidden)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) Conflict(body ...any) error {
	res := ctx.Status(http.StatusConflict)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx HttpResponse) NoContent() error {
	res := ctx.Status(http.StatusNoContent)
	ctx.bodyParse(res)
	return nil
}

func (ctx HttpResponse) Accepted(body ...any) error {
	res := ctx.Status(http.StatusAccepted)
	ctx.bodyParse(res, body...)
	return nil
}
