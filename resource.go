package goa

// Resource definitions describe REST resources exposed by the application API.
// They can be versioned so that multiple versions can be exposed (usually for
// backwards compatibility). Clients specify the version they want to use
// through the "X-API-VERSION" request header or "api_version" query string
// parameter. If an api version is specified in the resource definition then
// clients must specify the version header or get back a response with status
// code 404.
//
// The definition also includes a description, a route prefix, a media type and
// action definitions.
// The route prefix is the common path under which all resource actions are
// located. The complete URL for an action is application path + controller
// prefix + resource prefix + action path.
// The media type describes the fields of the resource (see media_type.go). The
// resource actions may define a different media type for their responses,
// typically the "index" and "show" actions re-use the resource media type.
// Action definitions list all the actions supported by the resource (both CRUD
// and other actions), see the Action struct.
type Resource struct {
	Name        string
	Description string
	ApiVersion  string
	RoutePrefix string
	MediaType   MediaType
	Actions     map[string]Action
}

// Media types are used to define the content of controller action responses.
// They provide the API clients with a crisp definition of what an API returns.
// Conceptually media types define the "views" of resources, the media type
// content is described using a json schema.
//
// A media type also define views which provide different ways of rendering its
// properties. For example there could be a "tiny" view that only includes a few
// attributes (e.g. "name", "href") used for listing and an "extended" view that
// includes all the properties.
//
// Finally, a media type may define a `links` property to represent related
// resources. The links property type must be an object, each property
// corresponding to a link. The type of the links properties must be reference
// to media types that define a "link" view. The link view is a special view
// that is used when rendering links to the media type defining it.
//
// A media type may define both a property and a link with the same name, views
// may then decide to use one or another (or even both). When a link is defined
// with the same name as an attribute of the media type then it does not have to
// redefine any of the attribute field, they get "inherited" from the media type
// attribute.
type MediaType struct {
	Identifier  string     // HTTP media type identifier (http://en.wikipedia.org/wiki/Internet_media_type)
	Description string     // Description used for documentation
	Schema      JsonSchema // Actual media type definition
	Views       Views      // Media type views
}

// Views have a description and a list of property names
type View struct {
	Description string   // View description
	Properties  []string // Name of properties to include in view
}

// Collection of named Views
type Views map[string]View

// Actions describe operation that can be run on resources. They define a route
// which consists of one ore more pairs of HTTP verb and path.
// They also optionally define the action parameters
// (variables defined in the route path) and payload (request body content).
// Parameters and payload are described using attributes which may include
// validations. goa takes care of validating and coercing the parameters and
// payload fields (and returns a response with status code 400 and a description
// of the validation error in the body in case of failure).
//
// The Multipart field specifies whether the request body must
// (RequiresMultipart) or can (SupportsMultipart) use a multipart content type.
// Multipart requests can be used to implement bulk actions - for example bulk
// updates. Each part contains the payload for a single resource, the same
// payload that would be used to apply the action to that resource in a standard
// (non-multipart) request.
//
// Action definitions also specify the set of views supported by the action.
// Different views may render the media type differently (ommitting certain
// attributes or links, see media_type.go). As with filters the client specifies
// the view in the special "view" URL query string:
//
//  "?view=tiny"
//
// Finally, action definitions describe the set of potential responses they may
// return and for each response the status code, compulsory headers and a media
// type (if different from the resource media type).
type Action struct {
	Name        string
	Description string
	Route       Route
	Params      []NamedSchemas
	Payload     *JsonSchema
	Views       []string
	Responses   Responses
	Multipart   int
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
