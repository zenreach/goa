package goa

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"path"
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
	Mount(controller Controller, definition *Resource)
	// ServeHTTP() implements http.HandlerFunc
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	// Handler() returns a Negroni handler that wraps the application
	Handler() negroni.Handler
	// PrintRoutes prints application routes to stdout
	PrintRoutes()
}

// Internal struct holding application data
// Implements the Application interface
type app struct {
	router      *mux.Router
	basePath    string
	controllers map[string]*Controller
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
		logger := &negroni.Logger{log.New(os.Stdout, "[goa] ", 0)}
		n = negroni.New(negroni.NewRecovery(), logger, negroni.NewStatic(http.Dir("public")))
	} else {
		// Custom handlers
		n = negroni.New(handlers...)
	}
	a := &app{
		router:      router,
		basePath:    basePath,
		controllers: make(map[string]*Controller),
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
	resourcePath := path.Join(app.basePath, resource.RoutePrefix)
	if _, ok := app.controllers[resourcePath]; ok {
		panic(fmt.Sprintf("goa: %v - controller already mounted under %s (%v)", reflect.TypeOf(controller), resourcePath, reflect.TypeOf(controller)))
	}
	if _, err := url.Parse(resourcePath); err != nil {
		panic(fmt.Sprintf("goa: %v - invalid path specification '%s': %v", reflect.TypeOf(controller), resourcePath, err))
	}
	sub := app.router
	version := resource.ApiVersion
	if len(version) != 0 {
		route := app.router.Headers("X-Api-Version", version)
		sub = route.Subrouter()
	}
	finalizeResource(resource)
	app.routeMap.addRoutes(resource, controller)
	app.addHandlers(sub, resourcePath, resource, controller)
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
func (app *app) PrintRoutes() {
	app.routeMap.PrintRoutes(app.basePath)
}

// validateResource validates resource definition recursively
func validateResource(resource *Resource) error {
	mediaType := &resource.MediaType
	if mediaType.IsEmpty() {
		return nil
	}
	return mediaType.Model.Validate()
}

// finalizeResource links child action and response definitions back to resource definition
func finalizeResource(resource *Resource) {
	resource.pActions = make(map[string]*Action, len(resource.Actions))
	for an, action := range resource.Actions {
		responses := make(Responses, len(action.Responses))
		for n, r := range action.Responses {
			responses[n] = r
		}
		params := make(Params, len(action.Params))
		for n, p := range action.Params {
			params[n] = p
		}
		pPayload := &Payload{
			Attributes: action.Payload.Attributes,
			Blueprint:  action.Payload.Blueprint,
		}
		filters := make(Filters, len(action.Filters))
		for n, p := range action.Filters {
			filters[n] = p
		}
		resource.pActions[an] = &Action{
			Name:        an,
			Description: action.Description,
			Route:       action.Route,
			Multipart:   action.Multipart,
			Views:       action.Views,
			pParams:     &params,
			pPayload:    pPayload,
			pFilters:    &filters,
			pResponses:  &responses,
		}
	}
}

// Register HTTP handlers for all controller actions
func (app *app) addHandlers(router *mux.Router, resourcePath string, resource *Resource, controller Controller) {
	for name, action := range resource.pActions {
		name = strings.ToUpper(string(name[0])) + name[1:]
		for _, route := range action.Route.GetRawRoutes() {
			matcher := router.Methods(route[0])
			actionPath := path.Join(resourcePath, route[1])
			elems := strings.SplitN(actionPath, "?", 2)
			actionPath, queryString := elems[0], elems[1]
			matcher = matcher.Path(actionPath)
			if len(queryString) > 0 {
				query := strings.Split(queryString, "&")
				for _, q := range query {
					pair := strings.SplitN(q, "=", 2)
					matcher = matcher.Queries(pair[0], pair[1])
				}
			}
			matcher.HandlerFunc(actionHandlerFunc(name, action, controller))
		}
	}
}

// Single action handler
// All the logic lies in the actionHandler struct which implements the standard http.HandlerFunc
func actionHandlerFunc(name string, action *Action, controller Controller) http.HandlerFunc {
	// Use closure for great benefits: do not build new handler for every request
	handler, err := newActionHandler(name, action, controller)
	if err != nil {
		panic(fmt.Sprintf("goa: %s", err.Error()))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
