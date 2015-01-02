package goa

import (
	"path"
	"regexp"
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
	controller Controller
	actions    map[string]*compiledAction
	apiVersion string
	fullPath   string
	name       string
}

// A compiled action uses pointers to refer to its fields and has an associated
// full path and resource.
type compiledAction struct {
	action     *Action
	hasPayload bool              // true if action accepts a payload, false otherwise.
	resource   *compiledResource // Parent resource definition
	routes     []*compiledRoute  // Base URI to action including app base path and resource route prefix
}

// A compiled route is the full url to an action request and its associated HTTP
// verb.
type compiledRoute struct {
	verb             string // One of "GET", "POST", "PUT", "DELETE" etc.
	path             string
	capturePositions map[string]int
}

// compileResource creates a compiled resource from a resource declaration.
// The compiled resource uses pointers to refer to the various resource fields
// instead of the values used for the DSL. Compiled resources also link actions
// back to their parent resources and contain fields that hold other
// pre-computed information like the action paths.
func compileResource(resource *Resource, controller Controller, appPath string) *compiledResource {
	resourcePath := appPath
	if len(resource.RoutePrefix) > 0 {
		resourcePath = path.Join(resourcePath, resource.RoutePrefix)
	}
	compiled := &compiledResource{
		controller: controller,
		apiVersion: resource.ApiVersion,
		fullPath:   resourcePath,
		name:       resource.Name,
	}
	compiled.actions = make(map[string]*compiledAction, len(resource.Actions))
	reg := regexp.MustCompile("(.*)(\\+.*)")
	for an, action := range resource.Actions {
		responses := make(Responses, len(action.Responses))
		for n, r := range action.Responses {
			r.resource = resource
			r.name = n
			if r.MediaType.Identifier == "Resource" {
				r.MediaType = resource.MediaType
			} else if r.MediaType.Identifier == "ResourceCollection" {
				r.MediaType = resource.MediaType
				// The below may need tweaking, for now just insert "+collection"
				// in the media type identifier.
				id := r.MediaType.Identifier
				if reg.MatchString(id) {
					r.MediaType.Identifier = reg.ReplaceAllString(id, "$1+collection$2")
				} else {
					r.MediaType.Identifier = id + "+collection"
				}
				r.MediaType.Description += " (collection)"
			}
			responses[n] = r
		}
		params := make(Params, len(action.Params))
		for n, p := range action.Params {
			params[n] = p
		}
		payload := Payload{
			Attributes: action.Payload.Attributes,
			Blueprint:  action.Payload.Blueprint,
		}
		filters := make(Filters, len(action.Filters))
		for n, p := range action.Filters {
			filters[n] = p
		}
		routes := action.Route.GetRawRoutes()
		cRoutes := make([]*compiledRoute, len(routes))
		hasPayload := len(action.Payload.Attributes) > 0
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
			if hasPayload {
				startPos = 2
			}
			for i, m := range matches {
				positions[m[1]] = i + startPos
			}
			cRoutes[i] = &compiledRoute{r[0], actionPath, positions}
		}
		copy := Action{
			Name:      an,
			Multipart: action.Multipart,
			Views:     action.Views,
			Params:    params,
			Payload:   payload,
			Filters:   filters,
			Responses: responses,
		}
		compiled.actions[an] = &compiledAction{
			action:     &copy,
			hasPayload: hasPayload,
			resource:   compiled,
			routes:     cRoutes,
		}
	}

	return compiled
}
