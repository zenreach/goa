package goa

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

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
	Mount(*design.Resource, *Handler) error
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
// goa calls the handler to process requests made to actions of the resource it implements.
// Mount validates that the given handler implements the given resource actions.
func (app *app) Mount(r *design.Resource, h *Handler) error {
	if h == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	v := reflect.ValueOf(h)
	for name, action := range r.Actions {
		methName := name
		meth := v.MethodByName(methName)
		if !meth.IsValid() {
			methName = inflect.Camelize(name)
			meth = v.MethodByName(methName)
			if !meth.IsValid() {
				return fmt.Errorf("handler must implement %s or %s", name, methName)
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
			return fmt.Errorf("invalid number of parameters for %s, expected %d, got %d",
				methName, len(paramTypesInOrder), t.NumIn())
		}
		for i := 0; i < t.NumIn(); i++ {
			at := t.In(i)
			if err := paramTypesInOrder[i].CanLoad(at, ""); err != nil {
				return fmt.Errorf("Incorrect type for parameter #%d of %s, expected type to be compatible with %v, got %v (%s)",
					i+1, methName, at, paramTypesInOrder[i].Name(), err)
			}
		}

	}
	handlers[r.Name] = h
	return nil
}

// WriteRaml writes the RAML representation of the API to the given writer.
// see http://raml.org
func (app *app) WriteRaml(w io.Writer) {
}
