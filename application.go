package goa

import (
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Public interface of a goa application
type Application interface {
	// goa applications implement http.Handler.
	http.Handler
	// Mount adds a HTTP handler to the application.
	// The handler is used to handle all routes under the given path prefix.
	Mount(pathPrefix string, handler http.Handler)
	// ServeFiles serves files from the given file system root.
	// The path must end with "/*filepath", files are then served from the local
	// path /defined/root/dir/*filepath.
	// For example if root is "/etc" and *filepath is "passwd", the local file
	// "/etc/passwd" would be served.
	// To use the operating system's file system implementation,
	// use http.Dir:
	//     app.ServeFiles("/src/*filepath", http.Dir("/var/www"))
	ServeFiles(path string, root http.FileSystem)
	// WriteRaml writes the RAML representation of the API, see http://raml.org.
	WriteRaml(io.Writer)
}

type Handler func(http.ResponseWriter, *http.Request)

// Internal application data structure
type app struct {
	*httprouter.Router
	Name        string
	Description string
}

// New creates a new goa application.
func New(name, desc string) Application {
	router := httprouter.New()
	app := app{Router: router, Name: name, Description: desc}
	return &app
}

// Mount adds a handler to the application.
func (app *app) Mount(pathPrefix string, handler http.Handler) {
	app.Router.Handler("GET", pathPrefix, handler)
	app.Router.Handler("HEAD", pathPrefix, handler)
	app.Router.Handler("OPTIONS", pathPrefix, handler)
	app.Router.Handler("POST", pathPrefix, handler)
	app.Router.Handler("PUT", pathPrefix, handler)
	app.Router.Handler("PATCH", pathPrefix, handler)
	app.Router.Handler("DELETE", pathPrefix, handler)
}

// WriteRaml writes the RAML representation of the API to the given writer.
// see http://raml.org
func (app *app) WriteRaml(w io.Writer) {
}
