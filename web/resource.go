package web

import (
	"net/http"
)

type ResourceHandler interface {
	Index(http.ResponseWriter, *http.Request)
	Show(http.ResponseWriter, *http.Request)
	Store(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Destroy(http.ResponseWriter, *http.Request)
}
