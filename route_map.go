package goa

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"reflect"
)

// routeData holds the route fields
type routeData struct {
	version    string
	verb       string
	path       string
	action     string
	controller string
}

// The RouteMap type exposes two public methods WriteRoutes and PrintRoutes that can be called to print the routes
// for all mounted resource actions.
type RouteMap []*routeData

// Sorted map by action
type byAction RouteMap

func (a byAction) Len() int           { return len(a) }
func (a byAction) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAction) Less(i, j int) bool { return (*a[i]).action < (*a[j]).action }

// WriteRoutes writes routes table to given io writer
func (m *RouteMap) WriteRoutes(writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Verb", "Path", "Action", "Controller", "Version"})
	for _, r := range *m {
		table.Append([]string{r.verb, r.path, r.action, r.controller, r.version})
	}
	table.Render()
}

// PrintRoutes prints routes to stdout
func (m *RouteMap) PrintRoutes() {
	m.WriteRoutes(os.Stdout)
}

// addRoutes records routes for given resource definition
func (m *RouteMap) addRoutes(resource *Resource, controller Controller) {
	for _, action := range resource.pActions {
		m.addRoute(resource, action, controller)
	}
}

// addRoute records a single route
func (m *RouteMap) addRoute(resource *Resource, action *Action, controller Controller) {
	prefix := resource.RoutePrefix
	if len(prefix) == 0 {
		prefix = "/"
	}
	for _, route := range action.Route.GetRawRoutes() {
		path := route[1]
		if len(path) > 0 {
			if string(path[0]) != "/" && string(prefix[len(prefix)-1]) != "/" {
				path = "/" + path
			} else if string(path[0]) == "/" && string(prefix[len(prefix)-1]) == "/" {
				if len(path) > 1 {
					path = path[1 : len(path)-1]
				} else {
					path = ""
				}
			}
		}
		r := routeData{
			version:    resource.ApiVersion,
			verb:       route[0],
			path:       prefix + path,
			action:     action.Name,
			controller: fmt.Sprintf("%v", reflect.TypeOf(controller)),
		}
		*m = append(*m, &r)
	}
}
