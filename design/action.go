package design

import (
	"fmt"
	"regexp"
)

// A resource action
// Defines an HTTP endpoint and the shape of HTTP requests and responses made to
// that endpoint.
// The shape of requests is defined via "parameters", there are path parameters
// (i.e. portions of the URL that define parameter values), query string
// parameters and a payload parameter (request body).
type Action struct {
	Name        string       // Action name, e.g. "create"
	Description string       // Action description, e.g. "Creates a task"
	HttpMethod  string       // HTTP method, e.g. "POST"
	Path        string       // HTTP URL suffix (appended to parent resource path)
	Responses   []*Response  // Set of possible response definitions
	PathParams  ActionParams // Path parameters if any
	QueryParams ActionParams // Query string parameters if any
	Payload     *Blueprint   // Payload blueprint (request body) if any
}

// Get initializes the action HTTP method to GET and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Get(path string) *Action {
	return a.method("Get", path)
}

// Post initializes the action HTTP method to POST and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Post(path string) *Action {
	return a.method("Post", path)
}

// Put initializes the action HTTP method to PUT and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Put(path string) *Action {
	return a.method("Put", path)
}

// Patch initializes the action HTTP method to PATCH and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Patch(path string) *Action {
	return a.method("Patch", path)
}

// Delete initializes the action HTTP method to DELETE and sets the path with the
// value passed as argument.
// It returns the action so that it can be chained with other setter methods.
// The path may define path parameters by prefixing URL elements with ':', e.g.:
//   "/tasks/:id"
func (a *Action) Delete(path string) *Action {
	return a.method("Delete", path)
}

// Respond adds a new action response using the given media type and a
// status code of 200.
func (a *Action) Respond(media *MediaType) *Response {
	r := Response{Status: 200, MediaType: media}
	a.Responses = append(a.Responses, &r)
	return &r
}

// RespondNoContent adds a new action response with no media type and a status
// code of 204.
func (a *Action) RespondNoContent() *Response {
	r := Response{Status: 204}
	a.Responses = append(a.Responses, &r)
	return &r
}

// WithParam creates a new query string parameter and returns it.
func (a *Action) WithParam(name string) *ActionParam {
	var param = &ActionParam{Name: name, Type: String}
	a.QueryParams[name] = param
	return param
}

// WithPayload sets the request payload type.
func (a *Action) WithPayload(payload *Blueprint) *Action {
	a.Payload = payload
	return a
}

// Regular expression used to capture path parameters
var pathRegex = regexp.MustCompile("/:([^/]+)")

// Internal helper method that sets HTTP method, path and path params
func (a *Action) method(method, path string) *Action {
	a.HttpMethod = method
	a.Path = path
	var matches = pathRegex.FindAllStringSubmatch(path, -1)
	a.PathParams = make(map[string]*ActionParam, len(matches))
	for _, m := range matches {
		a.PathParams[m[1]] = &ActionParam{Name: m[1]}
	}
	return a
}

// Validates that action definition is consistent: parameters have unique names, has at least one
// response.
func (a *Action) validate() error {
	if a.Name == "" {
		return fmt.Errorf("Action name cannot be empty")
	}
	if len(a.Responses) == 0 {
		return fmt.Errorf("Action %s has no response defined")
	}
	for _, p := range a.PathParams {
		for _, q := range a.QueryParams {
			if p.Name == q.Name {
				return fmt.Errorf("Action has both path parameter and query parameter named %s",
					p.Name)
			}
		}
	}
	return nil
}

// Generate go signature of controller function that can implement action.
//
func (a *Action) signature() string {
	return ""
}
