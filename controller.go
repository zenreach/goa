package goa

import (
	"fmt"
	"reflect"

	"bitbucket.org/pkg/inflect"
	"github.com/raphael/goa/design"
)

// A controller is an internal data structure that keeps track of mounted resources and their
// corresponding handler providers.
type controller struct {
	resource *design.Resource // Resource being implemented
	provider HandlerProvider  // Handler factory
	paramPos map[string]int   // Maps path and query parameter names to action argument positions
}

// newController validates that the given handler provider produces handlers that implement
// the given resource and creates a new controller if that's the case or returns an error otherwise.
func newController(r *design.Resource, p HandlerProvider) (*controller, error) {
	handler := p(nil, nil)
	if handler == nil {
		return nil, fmt.Errorf("handler provider returns nil objects")
	}
	v := reflect.ValueOf(handler)
	for name, action := range r.Actions {
		methName := name
		meth := v.MethodByName(methName)
		if !meth.IsValid() {
			methName = inflect.Camelize(name)
			meth = v.MethodByName(methName)
			if !meth.IsValid() {
				return nil, fmt.Errorf("handler must implement %s or %s", name, methName)
			}
		}
		t := meth.Type()
		var paramTypesInOrder []design.DataType
		if action.Payload != nil {
			paramTypesInOrder = append(paramTypesInOrder, action.Payload)
		}
		for _, p := range action.PathParams {
			paramTypesInOrder = append(paramTypesInOrder, p.Type)
		}
		for _, p := range action.QueryParams {
			paramTypesInOrder = append(paramTypesInOrder, p.Type)
		}
		if len(paramTypesInOrder) != t.NumIn() {
			return nil, fmt.Errorf("invalid number of parameters for %s, expected %d, got %d",
				methName, len(paramTypesInOrder), t.NumIn())
		}
		for i := 0; i < t.NumIn(); i++ {
			at := t.In(i)
			if err := paramTypesInOrder[i].CanLoad(at, ""); err != nil {
				return nil, fmt.Errorf("Incorrect type for parameter #%d of %s, expected type to be compatible with %v, got %v (%s)",
					i+1, methName, at, paramTypesInOrder[i].Name(), err.Error())
			}
		}

	}
	c := controller{resource: r, provider: p}
	return &c, nil
}
