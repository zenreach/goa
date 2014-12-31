package goa

import (
	"fmt"
	"strings"
)

// Resource definitions describe REST resources exposed by the application API.
// They can be versioned so that multiple versions can be exposed (usually for backwards compatibility). Clients
// specify the version they want to use through the X-API-VERSION request header. If an api version is specified in the
// resource definition then clients must specify the version header or get back a response with status code 404.
//
// The definition also includes a description, a route prefix, a media type and action definitions.
// The route prefix is the common path under which all resource actions are located. The complete URL for an action
// is application path + controller prefix + resource prefix + action path.
// The media type describes the fields of the resource (see media_type.go). The resource actions may define a
// different media type for their responses, typically the "index" and "show" actions re-use the resource media type.
// Action definitions list all the actions supported by the resource (both CRUD and other actions), see the Action
// struct.
type Resource struct {
	Name        string
	Description string
	ApiVersion  string
	RoutePrefix string
	MediaType   MediaType
	Actions     map[string]Action
}

// Action definitions define a route which consists of one ore more pairs of HTTP verb and path. They also optionally
// define the action parameters (variables defined in the route path) and payload (request body content). Parameters
// and payload are described using attributes which may include validations. goa takes care of validating and coercing
// the parameters and payload fields (and returns a response with status code 400 and a description of the validation
// error in the body in case of failure).
//
// The Multipart field specifies whether the request body must (RequiresMultipart) or can (SupportsMultipart) use a
// multipart content type. Multipart requests can be used to implement bulk actions - for example bulk updates. Each
// part contains the payload for a single resource, the same payload that would be used to apply the action to that
// resource in a standard (non-multipart) request.
//
// Action definitions may also specify a list of supported filters - for example an index action may support filtering
// the list of results given resource field values. Filters are defined using attributes, they are specified by the
// client using the special "filters" URL query string, the syntax is:
//
//   "?filters[]=some_field==some_value&&filters[]=other_field==other_value"
//
// Filters are readily available to the action implementation after they have been validated and coerced by goa. The
// exact semantic is up to the action implementation.
//
// Action definitions also specify the set of views supported by the action. Different views may render the media type
// differently (ommitting certain attributes or links, see media_type.go). As with filters the client specifies the
// view in the special "view" URL query string:
//
//  "?view=tiny"
//
// Finally, action definitions describe the set of potential responses they may return and for each response the status
// code, compulsory headers and a media type (if different from the resource media type). These response definitions
// are named so that the action implementation can create a response from its definition name.
type Action struct {
	Name        string
	Description string
	Route       Route
	Params      Params
	Payload     Payload
	Filters     Filters
	Views       []string
	Responses   Responses
	Multipart   int
}

// DSL

type Params Attributes
type Payload Model
type Filters Attributes

// ValidateResponse checks that the response content matches one of the action response definitions if any
func (a *Action) ValidateResponse(res *standardResponse) error {
	if len(a.Responses) == 0 {
		return nil
	}
	errors := []string{}
	for _, r := range a.Responses {
		if err := r.Validate(res); err == nil {
			return nil
		} else {
			errors = append(errors, err.Error())
		}
	}
	msg := "Response %+v does not match any of action '%s' response" +
		" definitions:\n  - %s"
	return fmt.Errorf(msg, res, a.Name, strings.Join(errors, "\n  - "))
}

// Interface implemented by action route
type Route interface {
	GetRawRoutes() [][]string // Retrieve pair of HTTP verb and action path
}

// Possible values for the Action struct "Multipart" field
const (
	SupportsMultipart = iota // Action request body may use multipart content type
	RequiresMultipart        // Action request body must use multipart content type
)

// Map of action definitions keyed by action name
type Actions map[string]Action

// Map of response definitions keyed by response name
type Responses map[string]Response

// HTTP verbs enum type
type httpVerb string

//  Route struct
type singleRoute struct {
	Verb httpVerb // Route HTTP verb
	Path string   // Route path
}

// HTTP Verbs enum
const (
	options httpVerb = "OPTIONS"
	get     httpVerb = "GET"
	head    httpVerb = "HEAD"
	post    httpVerb = "POST"
	put     httpVerb = "PUT"
	delete_ httpVerb = "DELETE"
	trace   httpVerb = "TRACE"
	connect httpVerb = "CONNECT"
	patch   httpVerb = "PATCH"
)

// OPTIONS creates a route with OPTIONS verb and given path
func OPTIONS(path string) Route {
	return singleRoute{options, path}
}

// GET creates a route with GET verb and given path
func GET(path string) Route {
	return singleRoute{get, path}
}

// HEAD creates a route with HEAD verb and given path
func HEAD(path string) Route {
	return singleRoute{head, path}
}

// POST creates a route with POST verb and given path
func POST(path string) Route {
	return singleRoute{post, path}
}

// PUT creates a route with PUT verb and given path
func PUT(path string) Route {
	return singleRoute{put, path}
}

// DELETE creates a route with DELETE verb and given path
func DELETE(path string) Route {
	return singleRoute{delete_, path}
}

// TRACE creates a route with TRACE verb and given path
func TRACE(path string) Route {
	return singleRoute{trace, path}
}

// PATCH creates a route with PATCH verb and given path
func PATCH(path string) Route {
	return singleRoute{patch, path}
}

// A multi-route is an array of routes
type multiRoute []singleRoute

// Multi creates a multi-route from the given list of routes
func Multi(routes ...singleRoute) multiRoute {
	return multiRoute(routes)
}

// GetRawRoutes returns the pair of HTTP verb and path for the route
func (r singleRoute) GetRawRoutes() [][]string {
	return [][]string{{string(r.Verb), r.Path}}
}

// GetRawRoutes returns the list of pairs of HTTP verb and path for the multi-route
func (m multiRoute) GetRawRoutes() [][]string {
	routes := make([][]string, len(m))
	for _, r := range m {
		routes = append(routes, []string{string(r.Verb), r.Path})
	}
	return routes
}
