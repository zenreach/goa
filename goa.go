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

// A goa application fundamentally consists of a router and a set of controllers and resource definitions that get
// "mounted" under given paths (URLs). The router dispatches incoming requests to the appropriate controller.
// Goa applications are created via the NewApplication() factory method.
// Goa application can be run directly via the built-in ServeHTTP() function or used as Negroni middleware using
// the Handler() function.
type app struct {
	router      *mux.Router
	controllers map[string]Controller
	routeMap    *RouteMap
	handler     negroni.Handler
}

// Public interface of a goa application
type Application interface {
	// Mount a controller
	Mount(definition *Resource, controller Controller)
	// Goa apps implement the standard http.HandlerFunc
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	// PrintRoutes prints application routes to stdout
	PrintRoutes()
}

// A goa controller can be any type (it just needs to implement one function per action it exposes)
type Controller interface{}

// Create new goa application given a base path
func NewApplication(basePath string) Application {
	router := mux.NewRouter().PathPrefix(basePath).Subrouter()
	return &app{router: router, controllers: make(map[string]Controller), routeMap: new(RouteMap)}
}

// Mount controller under given application and path
// Note that this method will panic on error (e.g. if the path prefix is already in use)
// This is to make sure that the web app won't even start in case of a blatant error
func (app *app) Mount(resource *Resource, controller Controller) {
	if resource == nil {
		panic(fmt.Sprintf("goa: %v - missing resource", reflect.TypeOf(controller)))
	}
	if err := validateResource(resource); err != nil {
		panic(fmt.Sprintf("goa: %v - invalid resource: %s", reflect.TypeOf(controller), err.Error()))
	}
	path := resource.RoutePrefix
	if _, ok := app.controllers[path]; ok {
		panic(fmt.Sprintf("goa: %v - controller already mounted under %s (%v)", reflect.TypeOf(controller), path, reflect.TypeOf(controller)))
	}
	if _, err := url.Parse(path); err != nil {
		panic(fmt.Sprintf("goa: %v - invalid path specification '%s': %v", reflect.TypeOf(controller), path, err))
	}
	route := app.router.PathPrefix(path)
	version := resource.ApiVersion
	if len(version) != 0 {
		route = route.Headers("X-Api-Version", version)
	}
	sub := route.Subrouter()
	finalizeResource(resource)
	app.routeMap.addRoutes(resource, controller)
	app.addHandlers(sub, resource, controller)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (app *app) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := &negroni.Logger{log.New(os.Stdout, "[goa] ", 0)}
	n := negroni.New(negroni.NewRecovery(), logger, negroni.NewStatic(http.Dir("public")))
	n.Use(app.Handler())
	n.ServeHTTP(w, req)
}

// Handler() returns a negroni handler/middleware that runs the application
func (app *app) Handler() negroni.Handler {
	return negroni.Wrap(app.router)
}

// PrintRoutes prints application routes to stdout
func (app *app) PrintRoutes() {
	app.routeMap.PrintRoutes()
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
		pResponses := make(map[string]*Response, len(action.Responses))
		for rn, response := range action.Responses {
			pResponses[rn] = &Response{
				Description: response.Description,
				Status:      response.Status,
				MediaType:   response.MediaType,
				Location:    response.Location,
				Headers:     response.Headers,
				Parts:       response.Parts,
				resource:    resource,
			}
		}
		pParams := make(Params, len(action.Params))
		for n, p := range action.Params {
			pParams[n] = p
		}
		pPayload := &Payload{
			Attributes: action.Payload.Attributes,
			Blueprint:  action.Payload.Blueprint,
		}
		pFilters := make(Filters, len(action.Filters))
		for n, p := range action.Filters {
			pFilters[n] = p
		}
		resource.pActions[an] = &Action{
			Name:        an,
			Description: action.Description,
			Route:       action.Route,
			Multipart:   action.Multipart,
			Views:       action.Views,
			pParams:     &pParams,
			pPayload:    pPayload,
			pFilters:    &pFilters,
			pResponses:  pResponses,
		}
	}
}

// Register HTTP handlers for all controller actions
func (app *app) addHandlers(router *mux.Router, definition *Resource, controller Controller) {
	for name, action := range definition.pActions {
		name = strings.ToUpper(string(name[0])) + name[1:]
		for _, route := range action.Route.GetRawRoutes() {
			matcher := router.Methods(route[0])
			elems := strings.SplitN(route[1], "?", 2)
			path := elems[0]
			var query []string
			if len(elems) > 1 {
				query = strings.Split(elems[1], "&")
			}
			matcher = matcher.Path(path)
			for _, q := range query {
				pair := strings.SplitN(q, "=", 2)
				matcher = matcher.Queries(pair[0], pair[1])
			}
			matcher.HandlerFunc(requestHandlerFunc(name, action, controller))
		}
	}
}

// Single action handler
// All the logic lies in the RequestHandler struct which implements the standard http.HandlerFunc
func requestHandlerFunc(name string, action *Action, controller Controller) http.HandlerFunc {
	// Use closure for great benefits: do not build new handler for every request
	handler, err := newRequestHandler(name, action, controller)
	if err != nil {
		panic(fmt.Sprintf("goa: %s", err.Error()))
	}
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}
