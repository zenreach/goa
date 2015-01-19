package goa

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Public interface of a goa application.
// A goa application fundamentally consists of a router and a set of controllers
// that get "mounted" under given paths (URLs).
// The router dispatches incoming requests to the appropriate controller action.
// Goa applications are created via the `New()` factory function. They can be
// run directly via the `ServeHTTP()` method or as a Negroni
// (https://github.com/codegangsta/negroni) middleware via the `Handler()`
// method. All routes mounted on an application can be retrieved using the
// `Routes()` method.
type Application interface {
	// Mount() adds a resource to the application
	Mount(*Controller)
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
	router    *mux.Router
	basePath  string
	resources []Resource
	routeMap  *RouteMap
	n         *negroni.Negroni
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
		router:    router,
		basePath:  basePath,
		resources: make([]Resource),
		routeMap:  new(RouteMap),
	}
	n.Use(a.Handler())
	a.n = n
	return a
}

// Mount adds an API controller to the application under a given path.
// This method panics on error (e.g. if the controller path is already in use)
// to make sure that the app won't even start in case of a blatant error.
func (app *app) Mount(path string, controller *Controller) {
	if controller == nil {
		panic(fmt.Sprintf("goa: API controller mounted under \"%s\" cannot be null", path))
	}
	err, res := compileResource(controller)
	if err != nil {
		panic(fmt.Sprintf("goa: invalid API resource: %s", err.Error()))
	}
	if _, ok := app.resources[res.fullPath]; ok {
		panic(fmt.Sprintf("goa: API resource already mounted under %s (%v)", res.fullPath))
	}
	if _, err := url.Parse(res.fullPath); err != nil {
		panic(fmt.Sprintf("goa: invalid path specification '%s': %v", res.fullPath, err))
	}
	app.resources = append(app.resources, res)
	app.routeMap.addRoutes(res)
	router := app.router
	version := res.ApiVersion
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

// Register HTTP handlers for all controller actions
func (app *app) addHandlers(router *mux.Router, resource *Resource) {
	for name, action := range resource.Actions {
		name = strings.ToUpper(string(name[0])) + name[1:]
		for _, route := range action.Routes {
			matcher := router.Methods(route.Verb)
			elems := strings.SplitN(route.Path, "?", 2)
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
			matcher.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
			})
		}
	}
}
