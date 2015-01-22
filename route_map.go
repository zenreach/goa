package goa

import (
	"bytes"
	"github.com/olekukonko/tablewriter"
	"io"
	"log"
	"sort"
	"strings"
)

// The RouteMap type exposes two public methods WriteRoutes and PrintRoutes that
// can be called to print the routes for all mounted resource actions.
type RouteMap struct {
	BasePath string
	Routes   []*RouteData
}

// RouteData holds the route fields
type RouteData struct {
	Version    string
	Verb       string
	Path       string
	Action     string
	Controller string
}

// AddRoute
func (m *RouteMap) AddRoute(

// PrintRoutes returns a formatted string that represent the routes for human
// consumption.
func (m *RouteMap) PrintRoutes() string {
	var b bytes.Buffer
	m.WriteRoutes(&b)
	return strings.TrimSpace(b.String())
}

// Log routes using given logger (uses info level)
func (m *RouteMap) Log(log *log.Logger) {
	routes := strings.Split(m.PrintRoutes(), "\n")
	for _, route := range routes {
		log.Printf(route)
	}
}

// WriteRoutes writes routes table to given io writer
func (m *RouteMap) WriteRoutes(writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Verb", "Path", "Action", "Controller", "Version"})
	for _, r := range sort.Sort(byAction(m.Routes)) {
		table.Append([]string{r.Verb, r.Path, r.Action, r.Controller, r.Version})
	}
	table.Render()
}

// Factory method
func newRouteMap(base string) *RouteMap {
	return &RouteMap{BasePath: base}
}

// Sorted map by action
type byAction RouteMap

func (a byAction) Len() int      { return len(a.Routes) }
func (a byAction) Swap(i, j int) { a.Routes[i], a.Routes[j] = a.Routes[j], a.Routes[i] }
func (a byAction) Less(i, j int) bool {
	ri, rj := verbRank(a.Routes[i].Verb), verbRank(a.Routes[j].Verb)
	if ri == rj {
		return len(a.Routes[i].Path) < len(a.Routes[j].Path)
	}
	return ri < rj
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
