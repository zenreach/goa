package goa

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

// A goa application fundamentally consists of a router and a set of controllers and resource definitions that get
// "mounted" under given paths (URLs). The router dispatches incoming requests to the appropriate controller.
// Goa applications are created via the `NewApplication()` factory method.
type app struct {
	router      *mux.Router
	controllers map[string]Controller
	routeMap    *RouteMap
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
// This is to make sure that the web app won't even start in case of a blattent error
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
	version := resource.ApiVersion
	if len(version) == 0 {
		panic(fmt.Sprintf("goa: %v - missing resource version", reflect.TypeOf(resource)))
	}
	sub := app.router.PathPrefix(path).Headers("X-Api-Version", version).Subrouter()
	finalizeResource(resource)
	app.routeMap.addRoutes(resource, controller)
	addHandlers(sub, resource, controller)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (app *app) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	app.router.ServeHTTP(w, req)
}

// PrintRoutes prints application routes to stdout
func (app *app) PrintRoutes() {
	app.routeMap.PrintRoutes()
}

// validateResource validates resource definition recursively
func validateResource(resource *Resource) error {
	return resource.MediaType.Model.Validate()
}

// finalizeResource links child action and response definitions back to resource definition
func finalizeResource(resource *Resource) {
	resource.pActions = make(map[string]*Action, len(resource.Actions))
	for an, action := range resource.Actions {
		clone := Action{action.Name, action.Description, action.Route, action.Params, action.Payload, action.Filters,
			action.Views, action.Responses, action.Multipart, nil, nil}
		resource.pActions[an] = &clone
		clone.resource = resource
		clone.Name = an
		clone.pResponses = make(map[string]*Response, len(action.Responses))
		for rn, response := range clone.Responses {
			cloneRes := Response{response.Description, response.Status, response.MediaType, response.Location,
				response.Headers, response.Parts, nil}
			clone.pResponses[rn] = &cloneRes
			cloneRes.resource = resource
		}
	}
}

// Route handler
type handlerPath struct {
	path    string
	handler http.HandlerFunc
	route   *mux.Route
}

// Array of route handler that supports sorting
type byPath []*handlerPath

func (a byPath) Len() int           { return len(a) }
func (a byPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPath) Less(i, j int) bool { return (*a[i]).path > (*a[j]).path }

// Register HTTP handlers for all controller actions
func addHandlers(router *mux.Router, definition *Resource, controller Controller) {
	// First create all routes
	handlers := byPath{}
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
			if len(path) > 0 {
				matcher = matcher.Path(path)
			}
			for _, q := range query {
				pair := strings.SplitN(q, "=", 2)
				matcher = matcher.Queries(pair[0], pair[1])
			}
			handlers = append(handlers, &handlerPath{path, requestHandlerFunc(name, action, controller), matcher})
		}
	}
	// Then sort them by path length (longer first) before registering them so that for example
	//  "/foo/{id}" comes before "/foo" and is matched first. Ideally should be handled by gorilla...
	sort.Sort(byPath(handlers))
	for _, h := range handlers {
		h.route.HandlerFunc(h.handler)
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
