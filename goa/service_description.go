package main

import ()

// Generated service description
// Results of parsing user source code
type apiDescription struct {
	resources   map[string]*resourceDef
	controllers map[string]*controllerDef
	mediaTypes  map[string]*mediaTypeDef
}

// Factory method
func newServiceDescription() *apiDescription {
	return &apiDescription{
		resources:   map[string]*resource{},
		mediaTypes:  map[string]*mediaType{},
		controllers: map[string]*controller{},
	}
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
			if len(action.mediaType) > 0 {
				_, ok := a.mediaTypes[action.mediaType]
				if !ok {
					return fmt.Errorf("Missing media type definition "+
						"%s used by action %s of resource %s", mt, n, name)
				}
			}
		}

	}
	for name, controller := range a.controllers {
		res = controller.resource
		_, ok := a.resources[res]
		if !ok {
			return fmt.Errorf("Missing resource definition "+
				"%s used by controller %s", res, name)
		}
	}
	return nil
}

// Generate API code
func (a *apiDescription) generate(w io.Writer) *Report {
	report := new(Report)
	names := make([]string, len(a.resources))
	idx := 0
	for name, _ := range a.resources {
		names[idx] = name
		idx += 1
	}
	sort.Strings(names)
	for _, name := range names {
		resource, _ := a.resources[name]
		resource.generate()
		for _, action := range resource.actions {
			for _, resp := range action.responses {
				if len(resp.mediaType) > 0 {
					mt, _ := a.mediaTypes[resp.mediaType]
					if err := mt.generate(); err != nil {
						return err
					}
				}
			}
		}
		controller, ok := a.controllers[name]
		
	}
}
