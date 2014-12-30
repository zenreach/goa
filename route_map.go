package goa

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"reflect"
	"sort"
)

// routeData holds the route fields
type routeData struct {
	version    string
	verb       string
	path       string
	action     string
	controller string
}

// rank returns the index of an HTTP verb in the routes table
func verbRank(verb string) int {
	order := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}
	for r, v := range order {
		if v == verb {
			return r
		}
	}
	panic("goa: Unknown HTTP verb " + verb)
}

// The RouteMap type exposes two public methods WriteRoutes and PrintRoutes that can be called to print the routes
// for all mounted resource actions.
type RouteMap struct {
	basePath string
	routes   []*routeData
}

// Sorted map by action
type byAction RouteMap

func (a byAction) Len() int      { return len(a.routes) }
func (a byAction) Swap(i, j int) { a.routes[i], a.routes[j] = a.routes[j], a.routes[i] }
func (a byAction) Less(i, j int) bool {
	ri, rj := verbRank(a.routes[i].verb), verbRank(a.routes[j].verb)
	if ri == rj {
		return len(a.routes[i].path) < len(a.routes[j].path)
	}
	return ri < rj
}

// WriteRoutes writes routes table to given io writer
func (m *RouteMap) WriteRoutes(writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	sort.Sort(byAction(*m))
	table.SetHeader([]string{"Verb", "Path", "Action", "Controller", "Version"})
	for _, r := range m.routes {
		table.Append([]string{r.verb, r.path, r.action, r.controller, r.version})
	}
	table.Render()
}

// PrintRoutes prints routes to stdout
func (m *RouteMap) PrintRoutes(basePath string) {
	m.basePath = basePath
	m.WriteRoutes(os.Stdout)
}

// addRoutes records routes for given resource definition
func (m *RouteMap) addRoutes(resource *Resource, controller Controller) {
	for _, action := range resource.Actions {
		m.addRoute(resource, &action, controller)
	}
}

// addRoute records a single route
func (m *RouteMap) addRoute(resource *Resource, action *Action, controller Controller) {
	prefix := m.basePath + resource.RoutePrefix
	if len(prefix) == 0 {
		prefix = "/"
	}
	for _, route := range action.Route.GetRawRoutes() {
		path := route[1]
		if len(path) > 0 {
			if string(path[0]) != "/" &&
				string(path[0]) != "?" &&
				string(prefix[len(prefix)-1]) != "/" {
				path = "/" + path
			} else if string(path[0]) == "/" && string(prefix[len(prefix)-1]) == "/" {
				if len(path) > 1 {
					path = path[1 : len(path)-1]
				} else {
					path = ""
				}
			}
		}
		version := resource.ApiVersion
		if len(version) == 0 {
			version = "N/A"
		}
		r := routeData{
			version:    version,
			verb:       route[0],
			path:       prefix + path,
			action:     action.Name,
			controller: fmt.Sprintf("%v", reflect.TypeOf(controller).Elem()),
		}
		m.routes = append(m.routes, &r)
	}
}
