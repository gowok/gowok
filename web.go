package gowok

import (
	"net/http"

	"github.com/gowok/gowok/web"
	"github.com/ngamux/ngamux"
)

func Router() *ngamux.HttpServeMux {
	return web.Router()
}

func Server() *http.Server {
	return web.Server()
}
