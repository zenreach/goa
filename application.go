package goa

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/raphael/goa/design"
)

// Public interface of a goa application
type Application interface {
	// goa applications implement http.Handler.
	http.Handler
	// Mount adds a resource to the application.
	// The resource is first validated: its name must not be blank and all the actions must
	// be valid. An action is valid if its name is not blank, it has at least one response
	// defined and parameter names are unique.
	// goa calls the handler provider to process requests made to actions of the resource.
	Mount(*design.Resource, HandlerProvider) error
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

// A handler provider creates new handler objects given a http request and response writer.
// goa calls the handler provider associated with the resource that contains the action being
// requested.
type HandlerProvider func(http.ResponseWriter, *http.Request) *Handler

// Internal application data structure
type app struct {
	*httprouter.Router
	Name         string
	Description  string
	controllers  []*controller
	bootstrapper *bootstrapper
}

// New creates a new goa application.
func New(name, desc string) Application {
	router := httprouter.New()
	app := app{Router: router, Name: name, Description: desc}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, bootstrapFlag) {
			b, err := newBootstrapper(arg[len(bootstrapFlag):])
			if err != nil {
				app.fail(err)
			}
			app.bootstrapper = b
			break
		}
	}
	return &app
}

// Mount adds a resource to the application.
// The resource is first validated: its name must not be blank and all the actions must
// be valid. An action is valid if its name is not blank, it has at least one response
// defined and parameter names are unique.
// goa calls the handler provider to process requests made to actions of the resource.
func (app *app) Mount(r *design.Resource, p HandlerProvider) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if c, err := newController(r, p); err != nil {
		return err
	} else {
		app.controllers = append(app.controllers, c)
		if app.bootstrapper != nil {
			if err := app.bootstrapper.process(c); err != nil {
				app.fail(err)
			}
		}
	}
	return nil
}

// WriteRaml writes the RAML representation of the API to the given writer.
// see http://raml.org
func (app *app) WriteRaml(w io.Writer) {
}

// Helper function that prints error that happens during bootstrap and exits.
func (app *app) fail(err error) {
	app.bootstrapper.cleanup()
	fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %s\n", err.Error())
	os.Exit(1)
}
