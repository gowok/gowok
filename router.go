package gowok

import (
	"net/http"

	"github.com/gowok/gowok/router"
	"github.com/ngamux/ngamux"
)

func Router() *ngamux.HttpServeMux {
	return router.Router()
}

func Server() *http.Server {
	return router.Server()
}
