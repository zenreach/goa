package goa

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"bitbucket.org/pkg/inflect"

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
func (app *app) Mount(r *design.Resource, h *Handler) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := registerHandler(r, p); err != nil {
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
	fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %s\n", err)
	os.Exit(1)
}

// registerHandler validates that the given handler provider produces handlers that implement
// the given resource and creates a new controller if that's the case or returns an error otherwise.
func registerHandler(r *design.Resource, handler *Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	v := reflect.ValueOf(handler)
	for name, action := range r.Actions {
		methName := name
		meth := v.MethodByName(methName)
		if !meth.IsValid() {
			methName = inflect.Camelize(name)
			meth = v.MethodByName(methName)
			if !meth.IsValid() {
				return nil, fmt.Errorf("handler must implement %s or %s", name, methName)
			}
		}
		t := meth.Type()
		var paramTypesInOrder []design.DataType
		if action.Payload != nil {
			paramTypesInOrder = append(paramTypesInOrder, action.Payload)
		}
		for _, p := range action.PathParams {
			paramTypesInOrder = append(paramTypesInOrder, p.Type)
		}
		for _, p := range action.QueryParams {
			paramTypesInOrder = append(paramTypesInOrder, p.Type)
		}
		if len(paramTypesInOrder) != t.NumIn() {
			return nil, fmt.Errorf("invalid number of parameters for %s, expected %d, got %d",
				methName, len(paramTypesInOrder), t.NumIn())
		}
		for i := 0; i < t.NumIn(); i++ {
			at := t.In(i)
			if err := paramTypesInOrder[i].CanLoad(at, ""); err != nil {
				return nil, fmt.Errorf("Incorrect type for parameter #%d of %s, expected type to be compatible with %v, got %v (%s)",
					i+1, methName, at, paramTypesInOrder[i].Name(), err)
			}
		}

	}
	handlers[r.Name] = handler
	return nil
}
