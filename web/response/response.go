package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ngamux/ngamux"
)

type Response struct {
	*ngamux.Response
}

func New(w http.ResponseWriter) *Response {
	return &Response{ngamux.Res(w)}
}

func (ctx Response) bodyParse(res *ngamux.Response, body ...any) {
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

func (r *Response) ToHttp() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *Response) Download(filepath string) {
	if filepath == "" {
		_ = r.NotFound(fmt.Sprintf("file %s is not found", filepath))
		return
	}

	f, err := os.Open(filepath)
	if err != nil {
		_ = r.NotFound(fmt.Sprintf("file %s is not found", filepath))
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	r.Header(
		"Content-Disposition", "attachment; filename="+info.Name(),
		"Content-Type", "application/octet-stream",
	)

	_, err = io.Copy(r, f)
	if err != nil {
		return
	}
}

func (r *Response) Header(kv ...string) *Response {
	header := r.Response.Header()

	if len(kv) < 2 {
		return r
	}

	for i := 0; i <= len(kv)/2; i += 2 {
		if kv[i] == "" {
			continue
		}

		header.Set(kv[i], kv[i+1])
	}

	return r
}
