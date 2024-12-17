package gowok

import (
	"net/http"

	"github.com/go-openapi/spec"
	"github.com/ngamux/ngamux"
)

type HttpDocs struct {
	swagger *spec.Swagger
}

type HttpDocsItem struct {
	*spec.PathItemProps
	method string
	path   string
}

func NewHttpDocs(title, version string) *HttpDocs {
	swagger := spec.Swagger{
		VendorExtensible: spec.VendorExtensible{},
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Schemes: []string{"http", "ws"},
			Host:    "localhost:8080",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: version,
					Title:   title,
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{},
			},
		},
	}

	return &HttpDocs{&swagger}
}

func (docs *HttpDocs) New(description string, callback func(*spec.Operation)) func(ngamux.Route) {
	operation := spec.NewOperation(description)
	operation.Description = description
	item := spec.PathItemProps{}
	return func(route ngamux.Route) {
		if callback != nil {
			callback(operation)
		}

		if itemFound, ok := docs.swagger.Paths.Paths[route.Path]; ok {
			item = itemFound.PathItemProps
		}

		switch route.Method {
		case http.MethodGet:
			item.Get = operation
		case http.MethodPost:
			item.Post = operation
		case http.MethodPut:
			item.Put = operation
		case http.MethodHead:
			item.Head = operation
		case http.MethodPatch:
			item.Patch = operation
		case http.MethodDelete:
			item.Delete = operation
		case http.MethodOptions:
			item.Options = operation
		}
		docs.swagger.Paths.Paths[route.Path] = spec.PathItem{
			PathItemProps: item,
		}
	}
}

func (docs HttpDocs) ServeHTTP(rw http.ResponseWriter, r *http.Request) error {
	return ngamux.Res(rw).JSON(docs.swagger)
}
