package gowok

import (
	"net/http"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/ngamux/ngamux"
)

type HttpDocs struct {
	swagger *spec.Swagger
}

type HttpDocsItem struct {
	*spec.PathItemProps
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
			Definitions: spec.Definitions{},
		},
	}

	return &HttpDocs{&swagger}
}

func (docs *HttpDocs) New(description string, callback func(*HttpDocsOperation)) func(ngamux.Route) {
	operation := &HttpDocsOperation{spec.NewOperation(description)}
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
			item.Get = operation.Operation
		case http.MethodPost:
			item.Post = operation.Operation
		case http.MethodPut:
			item.Put = operation.Operation
		case http.MethodHead:
			item.Head = operation.Operation
		case http.MethodPatch:
			item.Patch = operation.Operation
		case http.MethodDelete:
			item.Delete = operation.Operation
		case http.MethodOptions:
			item.Options = operation.Operation
		}
		docs.swagger.Paths.Paths[route.Path] = spec.PathItem{
			PathItemProps: item,
		}
	}
}

type HttpDocsDefinition struct {
	Name       string
	Type       string
	Properties map[string]spec.Schema
	Example    any
}

func (docs *HttpDocs) AddDefinition(definput HttpDocsDefinition) spec.Ref {
	definitions := make(map[string]spec.Schema)
	for key, def := range docs.swagger.SwaggerProps.Definitions {
		definitions[key] = def
	}
	definition := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       []string{definput.Type},
			Properties: definput.Properties,
		},
		SwaggerSchemaProps: spec.SwaggerSchemaProps{},
	}
	if definput.Example != nil {
		definition.SwaggerSchemaProps.Example = definput.Example
	}
	definitions[definput.Name] = definition
	docs.swagger.SwaggerProps.Definitions = spec.Definitions(definitions)

	return spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + definput.Name)}
}

func (docs HttpDocs) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ngamux.Res(rw).JSON(docs.swagger)
}

type HttpDocsOperation struct {
	*spec.Operation
}

type HttpDocsParam struct {
	Name      string
	In        string
	Type      []string
	Required  bool
	SchemaRef spec.Ref
}

func (o HttpDocsOperation) AddParam(param HttpDocsParam) {
	schemaProps := spec.SchemaProps{
		Nullable: !param.Required,
		Type:     spec.StringOrArray(param.Type),
		Ref:      param.SchemaRef,
	}
	swaggerSchemaProps := spec.SwaggerSchemaProps{}

	var _param *spec.Parameter
	switch param.In {
	case "body":
		_param = spec.BodyParam(param.Name, &spec.Schema{
			SchemaProps:        schemaProps,
			SwaggerSchemaProps: swaggerSchemaProps,
		})
	case "path":
		_param = spec.PathParam(param.Name)
	}
	o.Operation.AddParam(_param)
}

type HttpDocsRespond struct {
	Produces  []string
	Responses map[int]spec.ResponseProps
}

func (o HttpDocsOperation) AddResponds(respond HttpDocsRespond) {
	o.WithProduces(respond.Produces...)
	for key, res := range respond.Responses {
		o.RespondsWith(key, &spec.Response{
			ResponseProps: res,
		})
	}
}
