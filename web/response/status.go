package response

import "net/http"

func (ctx Response) Ok(body ...any) error {
	res := ctx.Status(http.StatusOK)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) BadRequest(body ...any) error {
	res := ctx.Status(http.StatusBadRequest)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) Unauthorized(body ...any) error {
	res := ctx.Status(http.StatusUnauthorized)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) NotFound(body ...any) error {
	res := ctx.Status(http.StatusNotFound)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) InternalServerError(body ...any) error {
	res := ctx.Status(http.StatusInternalServerError)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) Created(body ...any) error {
	res := ctx.Status(http.StatusCreated)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) Forbidden(body ...any) error {
	res := ctx.Status(http.StatusForbidden)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) Conflict(body ...any) error {
	res := ctx.Status(http.StatusConflict)
	ctx.bodyParse(res, body...)
	return nil
}

func (ctx Response) NoContent() error {
	res := ctx.Status(http.StatusNoContent)
	ctx.bodyParse(res)
	return nil
}

func (ctx Response) Accepted(body ...any) error {
	res := ctx.Status(http.StatusAccepted)
	ctx.bodyParse(res, body...)
	return nil
}
