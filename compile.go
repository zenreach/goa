package goa

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// A compiled resource is an internal struct used by goa at runtime when
// dispatching requests. The idea is to do as much pre-processing as possible
// before the app is actually started so that it's as efficient as possible.
// Steps taken during "compilation" include:
//   - Storing fields required at runtime in pointers so they don't have to be
//     copied each time they are accessed.
//   - Linking actions back to their parent resources for lookup.
//   - Computing action paths.
type compiledResource struct {
	controller  Controller
	actions     map[string]*compiledAction
	apiVersion  string
	fullPath    string
	name        string
	description string
}

// A compiled action uses pointers to refer to its fields and has an associated
// full path and resource.
type compiledAction struct {
	name        string
	description string
	multipart   int
	views       []string
	params      Params
	filters     Filters
	payload     *Model              // non-nil if action accepts a payload
	resource    *compiledResource   // Parent resource definition
	routes      []*compiledRoute    // Base URI to action including app base path and resource route prefix
	responses   []*compiledResponse // Action responses
}

// A compiled response embeds the response name, a link back to the original
// response and the original resource. It also contains a validated media type.
type compiledResponse struct {
	name      string            // Name used by error messages and documentation
	resource  *Resource         // Parent resource definition
	response  *Response         // Original response
	mediaType *MediaType        // Validated media type
	parts     *compiledResponse // Response part definitions if any
}

// A compiled route is the full url to an action request and its associated HTTP
// verb.
type compiledRoute struct {
	verb             string // One of "GET", "POST", "PUT", "DELETE" etc.
	path             string
	capturePositions map[string]int
}

// Regexp used to match media type identifier suffix to insert "+collection"
var idSuffixMatcher *regexp.Regexp

// compileResource creates a compiled resource from a resource declaration.
// The compiled resource uses pointers to refer to the various resource fields
// instead of the values used for the DSL. Compiled resources also link actions
// back to their parent resources and contain fields that hold other
// pre-computed information like the action paths.
func compileResource(resource *Resource, controller Controller, appPath string) (*compiledResource, error) {
	if idSuffixMatcher == nil {
		idSuffixMatcher = regexp.MustCompile("(.*)(\\+.*)")
	}
	resourcePath := appPath
	if len(resource.RoutePrefix) > 0 {
		resourcePath = path.Join(resourcePath, resource.RoutePrefix)
	}
	compiled := &compiledResource{
		controller:  controller,
		apiVersion:  resource.ApiVersion,
		fullPath:    resourcePath,
		name:        resource.Name,
		description: resource.Description,
	}
	compiled.actions = make(map[string]*compiledAction, len(resource.Actions))
	for an, action := range resource.Actions {
		responses := []*compiledResponse{}
		for n, r := range action.Responses {
			cr, err := compileResponse(n, &r, resource)
			if err != nil {
				return nil, err
			}
			responses = append(responses, cr)
		}
		params := make(Params, len(action.Params))
		for n, p := range action.Params {
			params[n] = p
		}
		var payload *Model
		if len(action.Payload.Attributes) > 0 {
			var err error
			payload, err = NewModel(action.Payload.Attributes, action.Payload.Blueprint)
			if err != nil {
				return nil, err
			}
		}
		filters := make(Filters, len(action.Filters))
		for n, p := range action.Filters {
			filters[n] = p
		}
		routes := action.Route.GetRawRoutes()
		cRoutes := make([]*compiledRoute, len(routes))
		for i, r := range routes {
			actionPath := resourcePath
			if len(r[1]) > 0 {
				if string(r[1][0]) != "?" {
					actionPath = path.Join(actionPath, r[1])
				} else {
					actionPath += r[1]
				}
			}
			positions := make(map[string]int)
			rexp := regexp.MustCompile("{([^}]+)}")
			matches := rexp.FindAllStringSubmatch(actionPath, -1)
			startPos := 1
			if payload != nil {
				startPos = 2
			}
			for i, m := range matches {
				positions[m[1]] = i + startPos
			}
			cRoutes[i] = &compiledRoute{r[0], actionPath, positions}
		}
		compiled.actions[an] = &compiledAction{
			name:        an,
			description: action.Description,
			multipart:   action.Multipart,
			views:       action.Views,
			params:      action.Params,
			filters:     action.Filters,
			payload:     payload,
			resource:    compiled,
			routes:      cRoutes,
			responses:   responses,
		}
	}

	return compiled, nil
}

// Compile a response, check its validity
func compileResponse(n string, r *Response, resource *Resource) (*compiledResponse, error) {
	respCopy := Response{
		Description: r.Description,
		Status:      r.Status,
		MediaType:   r.MediaType,
		Location:    r.Location,
		Headers:     r.Headers,
		Parts:       r.Parts,
	}
	cr := compiledResponse{
		name:     n,
		response: &respCopy,
		resource: resource,
	}
	mediaType := r.MediaType
	if mediaType.Identifier == "Resource" {
		mediaType = resource.MediaType
	} else if mediaType.Identifier == "ResourceCollection" {
		mediaType = resource.MediaType
		id := mediaType.Identifier
		// The below may need tweaking, for now just insert "+collection"
		// in the media type identifier.
		if idSuffixMatcher.MatchString(id) {
			mediaType.Identifier = idSuffixMatcher.ReplaceAllString(id, "$1+collection$2")
		} else {
			mediaType.Identifier = id + "+collection"
		}
		mediaType.Description += " (collection)"
	}
	var model *Model
	if (&mediaType).IsEmpty() {
		model = &mediaType.Model
	} else {
		// Validate attributes with blueprint
		var err error
		model, err = NewModel(mediaType.Model.Attributes, mediaType.Model.Blueprint)
		if err != nil {
			return nil, err
		}
	}
	cr.mediaType = &MediaType{
		Identifier:  mediaType.Identifier,
		Description: mediaType.Description,
		Model:       *model,
		Views:       mediaType.Views,
	}
	if r.Parts != nil {
		var err error
		cr.parts, err = compileResponse(n, r.Parts, resource)
		if err != nil {
			return nil, err
		}
	}
	return &cr, nil
}

// ValidateResponse checks that the response content matches one of the action response definitions if any
func (a *compiledAction) ValidateResponse(res *standardResponse) error {
	if len(a.responses) == 0 {
		return nil
	}
	errors := []string{}
	for _, r := range a.responses {
		if err := r.Validate(res); err == nil {
			return nil
		} else {
			errors = append(errors, err.Error())
		}
	}
	msg := "Response %+v does not match any of action '%s' response" +
		" definitions:\n  - %s"
	return fmt.Errorf(msg, res, a.name, strings.Join(errors, "\n  - "))
}

// Validate validates a response against its definition.
// It returns an error if validation fails, nil otherwise.
func (c *compiledResponse) Validate(r *standardResponse) error {
	d := c.response
	if d.Status == 500 {
		return nil // Already an error, protect against infinite loops
	}
	if d.Status > 0 {
		if r.Status() != d.Status {
			return fmt.Errorf("Response '%s': Value of response status does not match response definition (value is '%v', definition's is '%v')",
				c.name, r.Status(), d.Status)
		}
	}
	header := r.header
	if len(d.Location) > 0 {
		val := header.Get("Location")
		if !d.matches(val, d.Location) {
			return fmt.Errorf("Response '%s': Value of response header Location does not match response definition (value is '%s', definition's is '%s')",
				c.name, val, d.Location)
		}
	}
	if len(d.Headers) > 0 {
		for name, value := range d.Headers {
			val := strings.Join(header[http.CanonicalHeaderKey(name)], ",")
			if !d.matches(val, value) {
				return fmt.Errorf("Response '%s': Value of response header %s does not match response definition (value is '%s', definition's is '%s')",
					c.name, name, val, value)
			}
		}
	}
	media_type := c.mediaType
	id := media_type.Identifier
	if len(id) > 0 {
		parsed, _, err := mime.ParseMediaType(id)
		if err != nil {
			return fmt.Errorf("Response '%s': Invalid media type identifier '%s': %s",
				c.name, id, err.Error())
		}
		val := strings.Join(header["Content-Type"], ",")
		if parsed != strings.ToLower(val) {
			return fmt.Errorf("Response '%s': Value of response header Content-Type does not match response definition (value is '%s', definition's is '%s')",
				c.name, val, parsed)
		}
	}
	if d.Parts != nil {
		for name, part := range r.parts {
			if err := c.parts.Validate(part); err != nil {
				return fmt.Errorf("Response '%s': Invalid response part %s, %s",
					c.name, name, err.Error())
			}
		}
	}
	return nil
}
