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
	f, err := os.Open(filepath)
	if err != nil {
		r.NotFound(fmt.Sprintf("file %s is not found", filepath))
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	r.Header().Set("Content-Disposition", "attachment; filename="+info.Name())
	r.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(r, f)
	if err != nil {
		return
	}
}
