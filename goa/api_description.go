package main

import (
	"fmt"
	"io"
	"sort"
)

// Generated service description
// Results of parsing user source code
type apiDescription struct {
	resources   map[string]*resourceDef
	controllers map[string]*controllerDef
	mediaTypes  map[string]*mediaTypeDef
}

// Factory method
func newApiDescription() *apiDescription {
	return &apiDescription{
		resources:   map[string]*resourceDef{},
		controllers: map[string]*controllerDef{},
		mediaTypes:  map[string]*mediaTypeDef{},
	}
}

// Add resource to description
func (a *apiDescription) addResource(r *resourceDef) error {
	if _, ok := a.resources[r.name]; ok {
		return fmt.Errorf("Duplicate resource definition for resource %s", r.name)
	}
	a.resources[r.name] = r
	return nil
}

// Add media type to description
func (a *apiDescription) addMediaType(m *mediaTypeDef) error {
	if _, ok := a.mediaTypes[m.identifier]; ok {
		return fmt.Errorf("Duplicate media type definition for media type with identifier %s", m.identifier)
	}
	a.mediaTypes[m.identifier] = m
	return nil
}

// Add controller to description
func (a *apiDescription) addController(c *controllerDef) error {
	if _, ok := a.controllers[c.resource]; ok {
		return fmt.Errorf("Duplicate controller definition for controller implementing resource %s", c.resource)
	}
	a.controllers[c.resource] = c
	return nil
}

// Validate consistency of description
func (a *apiDescription) validate() error {
	for name, resource := range a.resources {
		mt := resource.mediaType
		_, ok := a.mediaTypes[mt]
		if !ok {
			return fmt.Errorf("Missing media type definition "+
				"%s used by resource %s", mt, name)
		}
		for n, action := range resource.actions {
			for _, response := range action.responses {
				if len(response.mediaType) > 0 {
					_, ok := a.mediaTypes[response.mediaType]
					if !ok {
						return fmt.Errorf("Missing media type definition "+
							"%s used by action %s of resource %s", mt, n, name)
					}
				}
			}
		}

	}
	for name, controller := range a.controllers {
		res := controller.resource
		_, ok := a.resources[res]
		if !ok {
			return fmt.Errorf("Missing resource definition "+
				"%s used by controller %s", res, name)
		}
	}
	return nil
}

// Generate API code
func (a *apiDescription) generate(w io.Writer) errors {
	names := make([]string, len(a.resources))
	idx := 0
	for name, _ := range a.resources {
		names[idx] = name
		idx += 1
	}
	sort.Strings(names)
	for _, name := range names {
		resource, _ := a.resources[name]
		resource.generate(w)
		for _, action := range resource.actions {
			for _, resp := range action.responses {
				if len(resp.mediaType) > 0 {
					mt, _ := a.mediaTypes[resp.mediaType]
					mt.generate(w)
				}
			}
		}
		//	controller, ok := a.controllers[name]

	}
	return errors{}
}
