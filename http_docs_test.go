package gowok

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/gowok/should"
	"github.com/ngamux/ngamux"
)

func httpDocsInitialize(t *testing.T) *HttpDocs {
	should := should.New(t)
	input := &HttpDocs{
		Title:               "1",
		Version:             "2",
		Host:                "localhost",
		Description:         "4",
		TermsOfService:      "5",
		BasePath:            "6",
		ContactName:         "7",
		ContactURL:          "8",
		ContactEmail:        "9",
		LicenseName:         "10",
		LicenseURL:          "11",
		Schemes:             []string{},
		Consumes:            []string{},
		Produces:            []string{},
		Tags:                []spec.Tag{},
		SecurityDefinitions: map[string]*spec.SecurityScheme{},
	}

	docs := NewHttpDocs(*input)
	should.NotNil(docs)
	should.NotNil(docs.swagger)
	input.swagger = docs.swagger

	should.Equal(input, docs)
	return docs
}

func TestNewHttpDocs(t *testing.T) {
	httpDocsInitialize(t)
}

func TestHttpDocsNew(t *testing.T) {
	docs := httpDocsInitialize(t)
	docs.New("1", func(o *spec.Operation) {
		should.NotNil(t, o)
	})(ngamux.Route{
		Method: "GET",
		Path:   "/users",
	})

	should.NotNil(t, docs.swagger.Paths.Paths["/users"].Get)
}

func TestHttpDocsAddDefinition(t *testing.T) {
	docs := httpDocsInitialize(t)
	type user struct{ Email string }
	ref := docs.AddDefinition(user{})

	should.Equal(t, ref.Ref.String(), "#/definitions/user")

	type userT struct {
		Email string `json:"email"`
	}
	ref = docs.AddDefinition(userT{})
	should.Equal(t, ref.Ref.String(), "#/definitions/userT")
}
