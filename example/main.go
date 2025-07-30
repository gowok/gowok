package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gowok/gowok"
	"github.com/gowok/gowok/router"
	"github.com/ngamux/ngamux"
)

func main() {
	gowok.Get().Configures(run).Run()
}

type Item[T any] struct {
	Type    string
	Content T
}

func run(p *gowok.Project) {
	others := p.Config.Others["key"].(bool)
	fmt.Println(
		others,
		reflect.TypeOf(others),
	)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res := ngamux.Res(w)
		res.Header().Set("Transfer-Encoding", "chunked")
		j := json.NewEncoder(w)

		err := j.Encode(Item[int]{"header", 0})
		if err != nil {
			gowok.HttpBadRequest(res, err)
			return
		}

		for i := 0; i < 20; i++ {
			err := j.Encode(Item[int]{"body", i})
			if err != nil {
				gowok.HttpBadRequest(res, err)
				continue
			}
			w.(http.Flusher).Flush()
		}
	})
}
