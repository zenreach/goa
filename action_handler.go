package goa

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"strings"
)

// A action handler implements the standard http HandlerFunc method for a
// single controller action.
type actionHandler struct {
	route      *compiledRoute
	action     *compiledAction // Compiled action
	controller Controller      // Instance of controller
	actionName string          // Action name
}

// Factory method
func newActionHandler(name string, route *compiledRoute, action *compiledAction,
	controller Controller) (*actionHandler, error) {
	if err := validateAction(name, action, controller); err != nil {
		return nil, err
	}
	return &actionHandler{
		route:      route,
		action:     action,
		controller: controller,
		actionName: name,
	}, nil
}

// ServeHTTP implements the standard net/http HandlerFunc function.
// The steps involved here are:
//   1. Parse and validate request parameters if any
//   2. Parse and validate request payload (a.k.a. body) if any
//   3. Call controller method with resulting request struct
func (handler *actionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := &Request{
		Raw:            r,
		ResponseWriter: w,
		response:       new(standardResponse),
	}
	if params, err := handler.loadParams(r); err != nil {
		request.respondError(400, "InvalidParam", err)
		return
	} else {
		request.Params = params
	}
	hasPayload := handler.action.hasPayload
	if hasPayload {
		if payload, err := handler.loadPayload(r); err != nil {
			request.respondError(400, "InvalidPayload", err)
			return
		} else {
			request.Payload = payload
		}
	}
	ln := len(request.Params) + 1
	if hasPayload {
		ln += 1
	}
	args := make([]reflect.Value, ln)
	for n, p := range request.Params {
		args[handler.route.capturePositions[n]] = reflect.ValueOf(p)
	}
	args[0] = reflect.ValueOf(request)
	if hasPayload {
		args[1] = reflect.ValueOf(request.Payload)
	}
	meth := reflect.ValueOf(handler.controller).MethodByName(handler.actionName)
	meth.Call(args)
	request.sendResponse(handler.action)
}

// validateAction validates the action definition against the controller method
// It makes the following checks:
//   1. There is a method on the controller whose name matches the name of the
//      action.
//   2. Make sure that the number of arguments on the action method matches the
//      number of attributes defined in both the action definition parameters
//      and payload.
//   3. Make sure that the type of the method arguments match the type of the
//      attributes.
func validateAction(name string, ca *compiledAction, controller Controller) error {
	// 1. Make sure there is a method with the right name on the controller
	actionMethod := reflect.ValueOf(controller).MethodByName(name)
	if actionMethod.Kind() != reflect.Func {
		return fmt.Errorf("No method '%s' exposed by controller %v",
			name, reflect.TypeOf(controller).Elem())
	}

	// 2. Make sure all routes are consistent (define the same captures)
	if len(ca.routes) == 0 {
		return fmt.Errorf("Action %s of resource %s defines no route."+
			" Please make sure all actions define at least one route.",
			ca.action.Name, ca.resource.name)
	}
	firstCaptures := ca.routes[0].capturePositions
	numArgs := len(firstCaptures)
	for i, route := range ca.routes[1:] {
		if len(route.capturePositions) != numArgs {
			return fmt.Errorf("Route #%d of action %s of resource"+
				" %s defines %d captures but first route defines %s."+
				" Please make sure that all routes of a given action"+
				" define the same number of captures.",
				i+1, ca.action.Name, ca.resource.name,
				len(route.capturePositions), numArgs)
		}
		for cap, idx := range route.capturePositions {
			if fidx, ok := firstCaptures[cap]; !ok {
				return fmt.Errorf("Route #%d of action %s of"+
					" resource %s defines capture '%s' but first"+
					" route does not. Please make sure all routes"+
					" of a given action define the same captures.",
					i+1, ca.action.Name, ca.resource.name,
					cap)
			} else {
				if fidx != idx {
					return fmt.Errorf("Route #%d of action %s of"+
						" resource %s defines capture '%s' at index"+
						" %d but first route defines it at index"+
						" %d. Please make sure all routes of a"+
						" given action define the same captures.",
						i+1, ca.action.Name, ca.resource.name,
						cap, idx, fidx)
				}
			}
		}

	}

	// 3. Make sure it has the right number of arguments
	action := ca.action
	attributes := action.Params
	hasPayload := ca.hasPayload
	argCount := len(attributes) + 1
	if hasPayload {
		argCount += 1
	}
	mType := actionMethod.Type()
	if argCount != mType.NumIn() {
		msg := "Method '%s' of controller %v takes %d argument(s)" +
			" but action %s of resource %s defines %d parameter(s)"
		if hasPayload {
			msg += " and a payload"
		}
		msg += ". Please make sure the action method has the correct" +
			" number of arguments."
		return fmt.Errorf(msg, name, reflect.TypeOf(controller),
			mType.NumIn(), ca.action.Name, ca.resource.name,
			len(attributes))
	}

	// 4. Validate action method argument types
	// TBD
	return nil
}

// toString returns the string representation for a given attribute type
func toString(t Type) string {
	switch t.GetKind() {
	case TString:
		return "string"
	case TInteger:
		return "int"
	case TFloat:
		return "float64"
	case TBoolean:
		return "bool"
	case TTime:
		return "time.Time"
	case TComposite:
		return "*struct"
	case TCollection:
		return fmt.Sprintf("[]%s", toString(t.(*Collection).ElemType))
	case THash:
		return fmt.Sprintf("map[string]%s", toString(t.(*Hash).ElemType))
	}
	panic(fmt.Sprintf("Unknown type %v", t))
}

// loadParams loads the values from the request url and applies the validation
// rules defined in the action definition.
// Parameters are defined in the action definition path and query string.
func (handler *actionHandler) loadParams(request *http.Request) (map[string]interface{}, error) {
	vars := mux.Vars(request)
	params := make(map[string]interface{})
	for name, attr := range handler.action.action.Params {
		val, ok := vars[name]
		var value interface{}
		if !ok {
			if attr.Required {
				return nil, errors.New("Missing required param " + name)
			} else if attr.DefaultValue != nil {
				value = attr.DefaultValue
			}
		} else {
			var err error
			value, err = attr.Type.Load(val)
			if err != nil {
				return nil, fmt.Errorf("Cannot load param '%s': %s", name, err.Error())
			}
		}
		params[name] = value
	}
	return params, nil
}

// loadPayload loads the payload attribute values from the request body and
// apply the validation rules defined in the attributes.
// This function supports loading form encoded, multi-part form encoded and
// JSON encoded bodies. The result is then loaded into an instance of the
// action payload blueprint.
func (handler *actionHandler) loadPayload(request *http.Request) (interface{}, error) {
	if request.ContentLength == 0 {
		return nil, nil
	}
	var parsed map[string]interface{}
	action := handler.action.action
	contentType := request.Header.Get("Content-Type")
	if strings.Contains(contentType, "form-urlencoded") {
		if err := request.ParseForm(); err != nil {
			return nil, fmt.Errorf("Failed to load form: %s", err.Error())
		}
		values := map[string][]string(request.PostForm)
		for name, attr := range action.Payload.Attributes {
			if val, err := handler.loadValue(name, values[name], &attr); err == nil {
				parsed[name] = val
			} else {
				return nil, fmt.Errorf("Failed to load form value %s: %s", name, err.Error())
			}
		}
	} else if strings.Contains(contentType, "multipart/form-data") {
		multipartReader, err := request.MultipartReader()
		if err != nil {
			return nil, fmt.Errorf("Failed to load multipart form: %s", err.Error())
		}
		form, err := multipartReader.ReadForm(int64(1024 * 1024 * 100))
		if err != nil {
			return nil, fmt.Errorf("Failed to parse multipart form: %s", err.Error())
		}
		for k, v := range form.Value {
			parsed[k] = v
		}
	} else if strings.Contains(contentType, "json") {
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&parsed)
		if err != nil {
			return nil, fmt.Errorf("Failed to load JSON: %s", err.Error())
		}
	} else if contentType == "" {
		return nil, errors.New("Empty Content-Type")
	} else {
		return nil, errors.New("Unsupported Content-Type")
	}

	for k, _ := range parsed {
		_, ok := action.Payload.Attributes[k]
		if !ok {
			return nil, fmt.Errorf("Unknown field '%s' in payload", k)
		}
	}
	payload, err := (*Model)(&action.Payload).Load(parsed)
	if err != nil {
		return nil, fmt.Errorf("Failed to load request payload: %s", err.Error())
	}

	return payload, nil
}

// loadValue loads a single value given a name, an incoming value and an
// attribute definition. This method both coerces the value if needed and
// validates it against the attribute definition. It returns the coerced value
// on success or an error on failure.
func (handler *actionHandler) loadValue(name string, raw interface{}, attr *Attribute) (interface{}, error) {
	var value interface{}
	if raw == nil {
		if attr.Required {
			return nil, errors.New("Missing required value " + name)
		} else if attr.DefaultValue != nil {
			value = attr.DefaultValue
		}
	} else {
		if reflect.TypeOf(raw).Kind() == reflect.Slice {
			// Coerce arrays into single elements if necessary
			// (to handle for example PostForm which always produces arrays for form values)
			arr := reflect.ValueOf(raw)
			if arr.Len() > 0 && reflect.TypeOf(attr.Type).Name() != "Collection" {
				raw = arr.Index(0)
			}
		}
		var err error
		value, err = attr.Type.Load(raw)
		if err != nil {
			return nil, fmt.Errorf("Cannot coerce %v %v into a %v (%v)", reflect.TypeOf(raw), raw, attr.Type, err)
		}
	}
	return &value, nil
}
