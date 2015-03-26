package goa

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
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
	Logger      io.Writer
	handler     http.Handler
	mux         *http.ServeMux
}

// New creates a new goa application.
func New(name, desc string, logger io.Writer) Application {
	mux := http.NewServeMux()
	handler := http.Handler(mux)
	if logger != nil {
		handler = handlers.LoggingHandler(logger, mux)
	}
	app := app{Logger: logger, Name: name, Description: desc, handler: handler, mux: mux}
	return &app
}

// ServerHTTP implements http.Handler.
// It uses a logging middleware if the logger given to New isn't nil.
func (app *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.handler.ServeHTTP(w, r)
}

// Mount registers the handler for the given prefix.
// If a handler already exists for prefix, Handle panics.
func (app *app) Mount(prefix string, handler http.Handler) {
	p := strings.TrimSuffix(prefix, "/")
	app.mux.Handle(p, handler)
	app.mux.Handle(p+"/", handler)
}
