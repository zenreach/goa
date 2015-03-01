package main

import (
	"fmt"
)

// Generated service description
// Results of parsing user source code
type apiDescription struct {
	resources   map[string]*ResourceDirective
	controllers map[string]*ControllerDirective
	mediaTypes  map[string]*MediaTypeDirective
}

// Factory method
func newApiDescription() *apiDescription {
	return &apiDescription{
		resources:   map[string]*ResourceDirective{},
		controllers: map[string]*ControllerDirective{},
		mediaTypes:  map[string]*MediaTypeDirective{},
	}
}

// Find controller associated with given resource
func (a *apiDescription) resourceCompiler(resource *ResourceDirective) (*ControllerDirective, error) {
	name := resource.name
	for _, c := range a.controllers {
		if c.resource == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("No controller implements resource %s", name)
}

// Add resource to description
func (a *apiDescription) addResource(r *ResourceDirective) error {
	if _, ok := a.resources[r.name]; ok {
		return fmt.Errorf("Duplicate resource definition for resource %s", r.name)
	}
	a.resources[r.name] = r
	return nil
}

// Add media type to description
func (a *apiDescription) addMediaType(m *MediaTypeDirective) error {
	if _, ok := a.mediaTypes[m.identifier]; ok {
		return fmt.Errorf("Duplicate media type definition for media type with identifier %s", m.identifier)
	}
	a.mediaTypes[m.identifier] = m
	return nil
}

// Add controller to description
func (a *apiDescription) addController(c *ControllerDirective) error {
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
			return fmt.Errorf("Missing media type with identifier "+
				"'%s' used by resource %s", mt, name)
		}
		for n, action := range resource.actions {
			for _, response := range action.responses {
				if len(response.mediaType) > 0 {
					_, ok := a.mediaTypes[response.mediaType]
					if !ok {
						return fmt.Errorf("Missing media type "+
							"with identifier %s "+
							"used by action %s of resource %s",
							mt, n, name)
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
	for _, resource := range a.resources {
		if _, err := a.resourceCompiler(resource); err != nil {
			return err
		}
	}
	return nil
}
