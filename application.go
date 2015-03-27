package goa

import (
	"net/http"
	"strings"
)

// Public interface of a goa application
type Application interface {
	// goa applications implement http.Handler.
	http.Handler
	// Mount registers the handler for the given prefix.
	// If a handler already exists for prefix, Handle panics.
	Mount(prefix string, handler http.Handler)
}

// Internal application data structure
type app struct {
	Name        string
	Description string
	mux         *http.ServeMux
}

// New creates a new goa application.
func New(name, desc string) Application {
	mux := http.NewServeMux()
	app := app{Name: name, Description: desc, mux: mux}
	return &app
}

// ServerHTTP implements http.Handler.
func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.mux.ServeHTTP(w, r)
}

// Mount registers the handler for the given prefix.
// If a handler already exists for prefix, Mount panics.
func (app *app) Mount(prefix string, handler http.Handler) {
	p := strings.TrimSuffix(prefix, "/")
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	app.mux.Handle(p, handler)
	app.mux.Handle(p+"/", handler)
}
