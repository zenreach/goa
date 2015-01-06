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
	method     reflect.Value   // Action controller method
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
		method:     reflect.ValueOf(controller).MethodByName(name),
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
	if handler.action.payload != nil {
		if payload, err := handler.loadPayload(r); err != nil {
			request.respondError(400, "InvalidPayload", err)
			return
		} else {
			request.Payload = payload
		}
	}
	ln := len(request.Params) + 1
	if handler.action.payload != nil {
		ln += 1
	}
	args := make([]reflect.Value, ln)
	for n, p := range request.Params {
		args[handler.route.capturePositions[n]] = reflect.ValueOf(p)
	}
	args[0] = reflect.ValueOf(request)
	if handler.action.payload != nil {
		args[1] = reflect.ValueOf(request.Payload)
	}
	handler.method.Call(args)
	request.sendResponse(handler.action)
}

// validateAction validates the action definition against the controller method.
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
			ca.name, ca.resource.name)
	}
	firstCaptures := ca.routes[0].capturePositions
	numArgs := len(firstCaptures)
	for i, route := range ca.routes[1:] {
		if len(route.capturePositions) != numArgs {
			return fmt.Errorf("Route #%d of action %s of resource"+
				" %s defines %d captures but first route defines %s."+
				" Please make sure that all routes of a given action"+
				" define the same number of captures.",
				i+1, ca.name, ca.resource.name,
				len(route.capturePositions), numArgs)
		}
		for cap, idx := range route.capturePositions {
			if fidx, ok := firstCaptures[cap]; !ok {
				return fmt.Errorf("Route #%d of action %s of"+
					" resource %s defines capture '%s' but first"+
					" route does not. Please make sure all routes"+
					" of a given action define the same captures.",
					i+1, ca.name, ca.resource.name,
					cap)
			} else {
				if fidx != idx {
					return fmt.Errorf("Route #%d of action %s of"+
						" resource %s defines capture '%s' at index"+
						" %d but first route defines it at index"+
						" %d. Please make sure all routes of a"+
						" given action define the same captures.",
						i+1, ca.name, ca.resource.name,
						cap, idx, fidx)
				}
			}
		}

	}

	// 3. Make sure it has the right number of arguments
	attributes := ca.params
	argCount := len(attributes) + 1
	if ca.payload != nil {
		argCount += 1
	}
	mType := actionMethod.Type()
	if argCount != mType.NumIn() {
		msg := "Method '%s' of controller %v takes %d argument(s)" +
			" but action %s of resource %s defines %d parameter(s)"
		if ca.payload != nil {
			msg += " and a payload"
		}
		msg += ". Please make sure the action method has the correct" +
			" number of arguments."
		return fmt.Errorf(msg, name, reflect.TypeOf(controller),
			mType.NumIn(), ca.name, ca.resource.name,
			len(attributes))
	}
	if len(firstCaptures) != len(attributes) {
		msg := "Action %s of resource %s defines %d parameter(s)" +
			" but route defines %d captures. Please make sure" +
			" these two numbers match."
		return fmt.Errorf(msg, ca.name, ca.resource.name,
			len(attributes), len(firstCaptures))
	}

	// 4. Validate action method argument types
	t := mType.In(0)
	var dummyRequest *Request
	if t != reflect.TypeOf(dummyRequest) {
		msg := "The type of the first argument of method '%s' of" +
			" controller %v is %v but should be *goa.Request."
		return fmt.Errorf(msg, name, reflect.TypeOf(controller), t)
	}
	if ca.payload != nil {
		t = mType.In(1).Elem()
		if err := ca.payload.CanLoad(t, ""); err != nil {
			msg := "The type of the second argument of method '%s'" +
				" of controller %v is %v but should be *%v (%s)."
			return fmt.Errorf(msg, name, reflect.TypeOf(controller),
				t, reflect.TypeOf(ca.payload.Blueprint),
				err.Error())
		}
	}
	if len(attributes) > 0 {
		for n, a := range attributes {
			idx, ok := firstCaptures[n]
			if !ok {
				msg := "Action %s of resource %s defines" +
					" parameter %s but there is no" +
					" corresponding capture defined in" +
					" the action route(s)."
				return fmt.Errorf(msg, ca.name,
					ca.resource.name, n)
			}
			t = mType.In(idx)
			if err := a.Type.CanLoad(t, ""); err != nil {
				msg := "Parameter %s of action %s of resource" +
					" %s is not compatible with the" +
					" corresponding argument of method '%s'" +
					" of controller %v - %s"
				return fmt.Errorf(msg, n, ca.name,
					ca.resource.name, name,
					reflect.TypeOf(controller), err.Error())
			}
		}
	}
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
	case TDateTime:
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
	for name, attr := range handler.action.params {
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
	action := handler.action
	payload := action.payload
	contentType := request.Header.Get("Content-Type")
	if strings.Contains(contentType, "form-urlencoded") {
		if err := request.ParseForm(); err != nil {
			return nil, fmt.Errorf("Failed to load request body: %s", err.Error())
		}
		values := map[string][]string(request.PostForm)
		for name, attr := range payload.Attributes {
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
		_, ok := payload.Attributes[k]
		if !ok {
			return nil, fmt.Errorf("Unknown field '%s' in payload", k)
		}
	}
	p, err := payload.Load(parsed)
	if err != nil {
		return nil, fmt.Errorf("Failed to load request payload: %s", err.Error())
	}

	return p, nil
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
