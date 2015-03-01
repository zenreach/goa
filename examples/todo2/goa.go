package goa

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Public interface of a goa application.
type Application interface {
	Mount(*goa.Controller)
	// ServeHTTP() implements http.HandlerFunc
	ServeHTTP(http.ResponseWriter, *http.Request)
	// Generate RAML representation for the API (http://raml.org)
	WriteRaml(io.Writer)
}

type App struct {
	Name string
	Description string
	BasePath string
}

// New creates a new goa application
func New(name string) Application {
	return &App{Name: name}
}

// WriteRaml returns the RAML representation of the API
// see http://raml.org
func (app *app) WriteRaml() string {
	return ""
}

func (app *app) NewResource() *goa.Resource {
}

type Resource struct {
	Name string
	Description string
	Version string
	MediaType *MediaType
	Actions []*Action
}

type MediaType struct {
	Identifier string
	Description string
	Attributes []*Attribute
}

type Attribute struct {
	Name string
	Type *Type
	Validators []TypeValidator
}

type TypeValidator interface {
	Validate(val interface{}) (TypeValidator, error)
}

type Action struct {
	Name string
	Description string
	Path string
	Responses []*ResponseSpec
}

type ResponseSpec struct {
	Name string
	Status int
	MediaType *MediaType
	Headers []*HeaderSpec
}

type HeaderSpec struct {
	Name string
	Pattern *regexp.Regexp
}

func (r *Resource) Action(name string) *goa.Action {
}

func (r *Resource) Show(path string) *goa.Action {
	var a = r.Action("show").Get(path)
	a.Respond(r.MediaType)
	return a
}