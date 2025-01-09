package gowok

import (
	"net/http"

	"github.com/ngamux/ngamux"
)

func HttpBadRequest(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusBadRequest)
	switch b := body.(type) {
	case string:
		res.Text(b)
	case *ValidationError, ValidationError:
		res.JSON(b)
	case error:
		res.Text(b.Error())
	default:
		res.JSON(b)
	}
}
func HttpUnauthorized(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusUnauthorized).Text("unauthorized")
}
func HttpNotFound(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusUnauthorized).Text("not found")
}
func HttpOk(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusOK)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpInternalServerError(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusInternalServerError)
	switch b := body.(type) {
	case string:
		res.Text(b)
	case error:
		res.Text(b.Error())
	default:
		res.JSON(b)
	}
}

func HttpCreated(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusCreated)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpForbidden(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusForbidden).Text("forbidden")
}

func HttpConflict(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusConflict)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpNoContent(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusNoContent).Text("")
}

func HttpAccepted(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusAccepted)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}
