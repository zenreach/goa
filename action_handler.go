package goa

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strings"
)

// A action handler implements the standard http HandlerFunc method for a
// single controller action.
type actionHandler struct {
	action       *Action        // Definition of action
	controller   Controller     // Instance of controller
	actionMethod reflect.Value  // Action method to be invoked
	actionName   string         // Action name
	positions    map[string]int // Position of arguments by attribute name
}

// Factory method
// This does 3 things:
// 1. Validate that the action method name and parameters match the action
//    definition
// 2. "Precompile" the action definition into a data structure used to
//    efficiently map values during requests
// 3. Build and return a pointer to a actionHandler struct
// Note that this is run once for each action at startup and never again once
// the service is running. So it's OK for this to be intensive and is a good
// place to do optimizations for later processing.
func newActionHandler(name string, action *Action, controller Controller) (*actionHandler, error) {
	if err := validateAction(name, action, controller); err != nil {
		return nil, err
	}
	actionMethod := reflect.ValueOf(controller).MethodByName(name)
	if positions, err := computeArgPos(action); err != nil {
		return nil, err
	} else {
		return &actionHandler{
			action:       action,
			controller:   controller,
			actionMethod: actionMethod,
			actionName:   name,
			positions:    positions,
		}, nil
	}
}

// ServeHTTP implements the standard net/http HandlerFunc function.
// The steps involved here are:
//   1. Parse and validate request parameters if any
//   2. Parse and validate request payload (a.k.a. body) if any
//   3. Call controller method with resulting request struct
func (handler *actionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	request := &Request{
		Raw:            req,
		ResponseWriter: w,
		response:       new(standardResponse),
	}
	if params, err := handler.loadParams(req); err != nil {
		request.respondError(400, "InvalidParam", err)
		return
	} else {
		request.Params = params
	}
	hasPayload := len(handler.action.pPayload.Attributes) > 0
	if hasPayload {
		if payload, err := handler.loadPayload(req); err != nil {
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
	args := make([]reflect.Value, ln, ln)
	for n, p := range request.Params {
		args[handler.positions[n]] = reflect.ValueOf(p)
	}
	args[0] = reflect.ValueOf(request)
	if hasPayload {
		args[1] = reflect.ValueOf(request.Payload)
	}
	handler.actionMethod.Call(args)
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
//      attributes. Note that the match cannot be done using names because
//      reflect doesn't let us look at the method argument names. Instead we
//      count that each argument type appears the right number times in the
//      attributes.
func validateAction(name string, action *Action, controller Controller) error {
	actionMethod := reflect.ValueOf(controller).MethodByName(name)
	if actionMethod.Kind() != reflect.Func {
		return fmt.Errorf("No method '%s' exposed by controller %v",
			name, reflect.TypeOf(controller).Elem())
	}
	attributes := *action.pParams
	hasPayload := len(action.pPayload.Attributes) > 0
	argCount := len(attributes) + 1
	if hasPayload {
		argCount += 1
	}
	mType := actionMethod.Type()
	if argCount != mType.NumIn() {
		msg := "The action method '%s' has %d argument(s) but the action definition has %d parameter(s)"
		if hasPayload {
			msg += " and a payload"
		}
		msg += ". Please make sure the action method has the correct number of arguments."
		return fmt.Errorf(msg, name, mType.NumIn(), len(attributes))
	}
	attrKinds := make(map[reflect.Kind]int)
	for _, v := range attributes {
		kind := toKind(v.Type)
		attrKinds[kind] += 1
	}
	if hasPayload {
		attrKinds[reflect.Struct] += 1
	}
	methKinds := make(map[reflect.Kind]int)
	for i := 0; i < mType.NumIn(); i++ {
		methKinds[mType.In(i).Kind()] += 1
	}
	for k, v := range attrKinds {
		if v != methKinds[k] {
			return fmt.Errorf("The action parameters and payload define %d attribute(s) of type %v but the action method '%s' defines %v arguments of that type. Make sure these two numbers match.",
				v, k, name, methKinds[k])
		}
	}
	return nil
}

// computeArgPos computes the method argument positions for each parameter.
// The position of parameters in the action method arguments matches their
// position in the action route.
func computeArgPos(action *Action) (map[string]int, error) {
	// TBD: Make sure all routes define the same captures
	routes := action.Route.GetRawRoutes()
	fullPath := path.Join(action.basePath, routes[0][1])
	positions := make(map[string]int)
	r := regexp.MustCompile("{([^}]+)}")
	matches := r.FindAllStringSubmatch(fullPath, -1)
	hasPayload := len(action.pPayload.Attributes) > 0
	startPos := 1
	if hasPayload {
		startPos = 2
	}
	for i, m := range matches {
		positions[m[1]] = i + startPos
	}
	return positions, nil
}

// toKind returns the reflect.Kind value for a given attribute type
func toKind(t Type) reflect.Kind {
	switch t.GetKind() {
	case TString:
		return reflect.String
	case TInteger:
		return reflect.Int
	case TFloat:
		return reflect.Float64
	case TBoolean:
		return reflect.Bool
	case TTime:
		return reflect.Struct
	case TComposite:
		return reflect.Struct
	case TCollection:
		return reflect.Slice
	case THash:
		return reflect.Map
	}
	panic(fmt.Sprintf("Unknown type %v", t))
}

// loadParams loads the values from the request url and applies the validation
// rules defined in the action definition.
// Parameters are defined in the action definition path and query string.
func (handler *actionHandler) loadParams(request *http.Request) (map[string]interface{}, error) {
	vars := mux.Vars(request)
	params := make(map[string]interface{})
	for name, attr := range *handler.action.pParams {
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
	contentType := request.Header.Get("Content-Type")
	if strings.Contains(contentType, "form-urlencoded") {
		if err := request.ParseForm(); err != nil {
			return nil, fmt.Errorf("Failed to load form: %s", err.Error())
		}
		values := map[string][]string(request.PostForm)
		for name, attr := range action.pPayload.Attributes {
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
		_, ok := action.pPayload.Attributes[k]
		if !ok {
			return nil, fmt.Errorf("Unknown field '%s' in payload", k)
		}
	}
	payload, err := (*Model)(action.pPayload).Load(parsed)
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
