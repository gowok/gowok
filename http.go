package gowok

import (
	"net/http"

	"github.com/gowok/gowok/web"
)

func HttpBadRequest(w http.ResponseWriter, body any) {
	web.NewResponse(w).BadRequest(body)
}

func HttpUnauthorized(w http.ResponseWriter) {
	web.NewResponse(w).Unauthorized("unauthorized")
}

func HttpNotFound(w http.ResponseWriter) {
	web.NewResponse(w).NotFound("not found")
}

func HttpOk(w http.ResponseWriter, body ...any) {
	web.NewResponse(w).Ok(body...)
}

func HttpInternalServerError(w http.ResponseWriter, body ...any) {
	web.NewResponse(w).InternalServerError(body...)
}

func HttpCreated(w http.ResponseWriter, body ...any) {
	web.NewResponse(w).Created(body...)
}

func HttpForbidden(w http.ResponseWriter) {
	web.NewResponse(w).Forbidden("forbidden")
}

func HttpConflict(w http.ResponseWriter, body ...any) {
	web.NewResponse(w).Conflict(body)
}

func HttpNoContent(w http.ResponseWriter) {
	web.NewResponse(w).NoContent()
}

func HttpAccepted(w http.ResponseWriter, body ...any) {
	web.NewResponse(w).Accepted(body...)
}
