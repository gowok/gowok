package gowok

import (
	"encoding/json"
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

func Handler(handler func(res *ngamux.Response, req *ngamux.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := ngamux.Res(w)
		err := handler(res, ngamux.Req(r))
		if err != nil {
			if gowokErr, ok := err.(errors.Error); ok {
				err = json.NewEncoder(res).Encode(gowokErr)
				if err == nil {
					return
				}
			}
			HttpInternalServerError(w, err)
		}
	}
}
