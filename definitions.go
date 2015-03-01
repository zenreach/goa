package goa

import "net/http"

// Resource definitions describe REST resources exposed by the application API.
// They can be versioned so that multiple versions can be exposed (usually for
// backwards compatibility). Clients specify the version they want to use
// through the "X-Api-Version" request header or "api_version" query string
// parameter. If an api version is specified in the resource definition then
// clients must specify the version or get back a response with status code 404.
//
// The definition also includes a description, a route prefix, a media type and
// action definitions.
// The route prefix is the common path under which all resource actions are
// located. The complete URL for an action is
// 'application path/route prefix/action path'.
// The media type describes the fields of the resource (see media_type.go). The
// resource actions may define a different media type for their responses,
// typically the "index" and "show" actions re-use the resource media type.
// Action definitions list all the actions supported by the resource (both CRUD
// and other actions), see the ActionDefinition struct.
type ResourceDefinition struct {
	Name        string
	Description string
	ApiVersion  string
	RoutePrefix string
	MediaType   MediaType
	Actions     map[string]*ActionDefinition
}

// Media types are used to define the content of controller action responses.
// They provide the API clients with a crisp definition of what an API returns.
// Conceptually media types define the "views" of resources, the media type
// content is described using a json schema.
//
// A media type also define views which provide different ways of rendering its
// properties. For example there could be a "tiny" view that only includes a few
// properties (e.g. "name" and "href") used for listing and an "extended" view
// that includes all the properties.
//
// A media type may define a `links` property to represent related resources.
// The links property type must be an object, each property corresponding to a
// link. The type of the links properties must be a reference to media types
// that define a "link" view. The link view is a special view that is used when
// rendering links to the media type defining it.
//
// A media type may define both a property and a link with the same name, views
// may then decide to use one or another (or even both). When a link is defined
// with the same name as an property of the media type then it does not have to
// redefine any of the property field, they get "inherited" from the media type
// property.
type MediaType struct {
	Identifier   string       // HTTP media type identifier (http://en.wikipedia.org/wiki/Internet_media_type)
	Description  string       // Description used for documentation
	Schema       string       // Actual media type definition as JSON schema
	Views        Views        // Media type views
	ViewMappings ViewMappings // Media type view mappings, see ViewMappings
}

// Collection of Views, each view lists the names of the properties it returns.
type Views map[string][]string

// View mappings give the name of the view to use to render a property that is
// itself a resource (or an array of resources) according to the view being used
// to render the overall resource.
// So for example a "blog" resource may contain a "posts" field and may define a
// view mapping that tells the renderer to use the "tiny" view to render the
// post resources when the "default" view is used to render the blog.
// The top level key is the name of the property, the value is a map that gives
// the name of the view used to render the embedded resource keyed by the name
// of the view used to render the overall resource.
type ViewMappings map[string]map[string]string

// Action definitions describe operation that can be run on resources. They
// define then HTTP method and path of the corresponding HTTP requests.
// They also optionally define the action parameters (variables defined in the
// route path) and payload (request body content).
// Parameters and payload are described using JSON schemas which may include
// validations. goa takes care of validating and coercing the parameters and
// payload fields (and returns a response with status code 400 and a description
// of the validation error in the body in case of failure).
//
// Action definitions also specify the set of views supported by the action.
// Different views may render the media type differently (ommitting certain
// attributes or links, see media_type.go). Clients specify the view in the
// special "view" URL query string, for example:
//
//  "?view=tiny"
//
// Action definitions may also describe the set of potential responses they
// return  and for each response the status code, compulsory headers and a media
// type  (if different from the resource media type).
// Finally, action definitions include the http HandlerFunc that provides the
// actual / implementation of the action.
type ActionDefinition struct {
	Name        string                      // Name of action
	Description string                      // Description used to generate documentation
	Method      string                      // HTTP method, one of "GET", "POST", etc.
	Path        string                      // Action path, relative to resource base path
	Params      map[string]string           // JSON schemas for parameters defined in URL path
	Queries     map[string]string           // JSON schemas for parameters defined in URL query
	Payload     string                      // JSON schema of action payload
	Views       []string                    // Supported views
	Responses   map[int]*ResponseDefinition // List of possible responses
	Handler     http.HandlerFunc            // Actual implementation
}

// Response definitions dictate the set of valid responses a given action may
// return. A response definition describes the response status code, media type
// and compulsory headers.
type ResponseDefinition struct {
	Status    int        // Response status code
	MediaType *MediaType // Response media type if any
	Headers   Headers    // Response header validations, enclose values in '/' for regexp behavior
}
