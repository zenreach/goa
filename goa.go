package goa

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

// Public interface of a goa application.
// A goa application fundamentally consists of a router and a set of controller
// and resource definition pairs that get "mounted" under given paths (URLs).
// The router dispatches incoming requests to the appropriate controller.
// Goa applications are created via the `New()` factory function. They can be
// run directly via the `ServeHTTP()` method or as a Negroni
// (https://github.com/codegangsta/negroni) middleware via the `Handler()`
// method. All routes mounted on an application can be printed using the
// `PrintRoutes()` method.
type Application interface {
	// Mount() adds a controller and associated resource to the application
	Mount(Controller, *Resource)
	// ServeHTTP() implements http.HandlerFunc
	ServeHTTP(http.ResponseWriter, *http.Request)
	// Handler() returns a Negroni handler that wraps the application
	Handler() negroni.Handler
	// Routes returns the application route map
	Routes() *RouteMap
}

// Internal struct holding application data
// Implements the Application interface
type app struct {
	router      *mux.Router
	basePath    string
	controllers map[string]*Controller
	resources   map[string]*compiledResource
	routeMap    *RouteMap
	n           *negroni.Negroni
}

// New creates a new goa application given a base path and an optional set of
// Negroni handlers (middleware).
func New(basePath string, handlers ...negroni.Handler) Application {
	router := mux.NewRouter()
	var n *negroni.Negroni
	if len(handlers) == 0 {
		// Default handlers a la "Negroni Classic()"
		logger := &negroni.Logger{log.New(os.Stdout, "[goa] ",
			log.Ldate|log.Lmicroseconds)}
		n = negroni.New(negroni.NewRecovery(), logger,
			negroni.NewStatic(http.Dir("public")))
	} else {
		// Custom handlers
		n = negroni.New(handlers...)
	}
	a := &app{
		router:      router,
		basePath:    basePath,
		controllers: make(map[string]*Controller),
		resources:   make(map[string]*compiledResource),
		routeMap:    new(RouteMap),
	}
	n.Use(a.Handler())
	a.n = n
	return a
}

// Mount adds a controller and associated resource to the application.
// The route to the controller is defined in the resource.
// This method panics on error (e.g. if the reource path prefix is already in
// use) to make sure that the app won't even start in case of a blatant error.
func (app *app) Mount(controller Controller, resource *Resource) {
	if resource == nil {
		panic(fmt.Sprintf("goa: %v - missing resource", reflect.TypeOf(controller)))
	}
	if err := validateResource(resource); err != nil {
		panic(fmt.Sprintf("goa: %v - invalid resource: %s", reflect.TypeOf(controller), err.Error()))
	}
	compiled, err := compileResource(resource, controller, app.basePath)
	if err != nil {
		panic(fmt.Sprintf("goa: %v - invalid resource: %s", reflect.TypeOf(controller), err.Error()))
	}
	if _, ok := app.controllers[compiled.fullPath]; ok {
		panic(fmt.Sprintf("goa: %v - controller already mounted under %s (%v)", reflect.TypeOf(controller), compiled.fullPath, reflect.TypeOf(controller)))
	}
	if _, err := url.Parse(compiled.fullPath); err != nil {
		panic(fmt.Sprintf("goa: %v - invalid path specification '%s': %v", reflect.TypeOf(controller), compiled.fullPath, err))
	}
	app.resources[resource.Name] = compiled
	app.routeMap.addRoutes(compiled, controller)
	router := app.router
	version := resource.ApiVersion
	if len(version) != 0 {
		route := app.router.Headers("X-Api-Version", version)
		router = route.Subrouter()
	}
	app.addHandlers(router, compiled, controller)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (app *app) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	app.n.ServeHTTP(w, req)
}

// Handler() returns a negroni handler/middleware that runs the application
func (app *app) Handler() negroni.Handler {
	return negroni.Wrap(app.router)
}

// PrintRoutes prints application routes to stdout
func (app *app) Routes() *RouteMap {
	return app.routeMap
}

// validateResource validates resource definition recursively
func validateResource(resource *Resource) error {
	mediaType := &resource.MediaType
	if mediaType.IsEmpty() {
		return nil
	}
	return mediaType.Model.Validate()
}

// Register HTTP handlers for all controller actions
func (app *app) addHandlers(router *mux.Router, resource *compiledResource, controller Controller) {
	for name, action := range resource.actions {
		name = strings.ToUpper(string(name[0])) + name[1:]
		for _, route := range action.routes {
			matcher := router.Methods(route.verb)
			elems := strings.SplitN(route.path, "?", 2)
			actionPath := elems[0]
			queryString := ""
			if len(elems) > 1 {
				queryString = elems[1]
			}
			matcher = matcher.Path(actionPath)
			if len(queryString) > 0 {
				query := strings.Split(queryString, "&")
				for _, q := range query {
					pair := strings.SplitN(q, "=", 2)
					matcher = matcher.Queries(pair[0], pair[1])
				}
			}
			// Use closure for great benefits: do not build new handler for every request
			handler, err := newActionHandler(name, route, action, controller)
			if err != nil {
				panic(fmt.Sprintf("goa: %s\nExpected signature:\n%s", err.Error(), expectedSignature(name, action, controller)))
			}
			matcher.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler.ServeHTTP(w, r) })
		}
	}
}

// String that represents action expected signature for error messages
func expectedSignature(name string, ca *compiledAction, controller Controller) string {
	prefix := fmt.Sprintf("func (c %v) %s(r *goa.Request", reflect.TypeOf(controller), name)
	args := []string{}
	if ca.payload != nil {
		args = []string{fmt.Sprintf("p *%v", reflect.TypeOf(ca.payload.Blueprint))}
	}
	for n, a := range ca.params {
		args = append(args, fmt.Sprintf("%s %s", n, toString(a.Type)))
	}
	if len(args) > 0 {
		return fmt.Sprintf("%s, %s)", prefix, strings.Join(args, ", "))
	} else {
		return fmt.Sprintf("%s)", prefix)
	}
}
