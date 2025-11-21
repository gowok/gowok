package request

import (
	"net/http"

	"github.com/ngamux/ngamux"
)

type Request struct {
	*ngamux.Request
	raw *http.Request
}

func New(r *http.Request) *Request {
	return &Request{ngamux.Req(r), r}
}

func (r *Request) ToHttp() *http.Request {
	return r.raw
}
