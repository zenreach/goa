package goa

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// The request handler is an internal object that handles a controller action request.
// It first validates the request then calls the controller action and finally validates the response.

// Controller actions take a `Request` interface as only parameter.
// The interface exposes methods to retrieve the coerced request parameters and payload attributes as well as the raw
// HTTP request object.
//
// The same interface also exposes methods to send the response back. There are a few ways to do this:
// - use `Respond()` to specify a response object. This object can be any object that implements the Response
// interface.
// - use `RespondEmpty()`, `RespondWithBody()` or `RespondWithHeader()` to send a response given its name and content
// the name of a response must match one of the action response definition names
// - use `RespondInternalError()` to return an error response (status 500)
//
// The Request interface also exposes a `ResponseBuilder()` method which given a response name returns an object that
// can be used to build the corresponding response (which can then be sent using the `Respond()` method described above)
type Request interface {
	Respond(response ResponseData)                                        // Send given response
	RespondEmpty(name string)                                             // Helper to respond with empty response
	RespondWithBody(name string, body interface{})                        // Helper to respond with body
	RespondWithHeader(name string, body interface{}, header *http.Header) // Helper to respond with body and headers
	RespondInternalError(body interface{})                                // Helper to respond with 500 and error message

	ResponseBuilder(name string) (ResponseBuilder, error) // Retrieve response builder to build more complex responses

	Param(name string) interface{}   // Retrieve parameter, requires type assertion before value can be used
	ParamString(name string) string  // Retrieve string parameter
	ParamInt(name string) int64      // Retrieve integer parameter
	ParamBool(name string) bool      // Retrieve boolean parameter
	ParamFloat(name string) float64  // Retrieve float parameter
	ParamTime(name string) time.Time // Retrieve time parameter

	Payload() interface{} // Retrieve payload, type is pointer to blueprint struct

	RawRequest() *http.Request // Underlying http request
}

// Internal request struct given to controller actions, implements Request interface
type requestData struct {
	responses   *map[string]*Response
	payload     interface{}
	params      *map[string]interface{}
	httpRequest *http.Request
	response    ResponseData
}

// ResponseBuilder returns a response builder corresponding to the response definition with given name.
// Call the `Response()` method on the ResponseBuildiner interface once the response has been initialized to send it
// via the `Request` interface `Respond()` method.
func (r *requestData) ResponseBuilder(responseName string) (ResponseBuilder, error) {
	if def, err := r.definition(responseName); err == nil {
		return &standardResponse{definition: def}, nil
	} else {
		return nil, err
	}
}

// Respond sets the response to be sent back to the client.
func (r *requestData) Respond(response ResponseData) {
	r.response = response
}

// RespondEmpty sets the response to be sent back to the client with an empty body.
// The response is initialized with the status code, media type and location header defined in the response definition
// with the given name.
func (r *requestData) RespondEmpty(name string) {
	r.initResponse(name, "", nil)
}

// RespondWithBody sets the response to be sent back to the client by serializing the given body object into JSON.
// The response is initialized with the status code, media type and location header defined in the response definition
// with the given name.
func (r *requestData) RespondWithBody(name string, body interface{}) {
	r.initResponse(name, body, nil)
}

// RespondWithHeader sets the response to be sent back to the client by serializing the given body object into JSON and
// using the given headers.
// The response is initialized with the status code, media type and location header defined in the response definition
// with the given name.
func (r *requestData) RespondWithHeader(name string, body interface{}, header *http.Header) {
	r.initResponse(name, body, header)
}

// RespondInternalError sets the response to be sent back to the client by serializing the given body object into JSON
// and using status code 500.
func (r *requestData) RespondInternalError(body interface{}) {
	r.Respond(&standardResponse{body: body})
}

// initResponse initializes the response from its definition name, given body and given custom headers.
func (r *requestData) initResponse(name string, body interface{}, header *http.Header) {
	if r.responses == nil {
		r.RespondInternalError("Request data not initialized")
	} else if def, ok := (*r.responses)[name]; !ok {
		r.RespondInternalError("Could not find response with name '" + name + "'")
	} else {
		if header == nil {
			header = &http.Header{}
		}
		if len(header.Get("Content-Type")) == 0 {
			id := def.MediaType.Identifier
			if len(id) == 0 {
				id = def.resource.MediaType.Identifier
			}
			if len(id) > 0 {
				header.Set("Content-Type", id)
			}
		}
		if len(header.Get("Location")) == 0 && len(def.Location) > 0 {
			header.Set("Location", def.Location)
		}
		r.Respond(&standardResponse{definition: def, status: def.Status, body: body, header: header})
	}
}

// definition returns the response definition with given name.
func (r *requestData) definition(name string) (*Response, error) {
	if r.responses == nil {
		return nil, errors.New("Request data not initialized")
	}
	if definition, ok := (*r.responses)[name]; ok {
		return definition, nil
	} else {
		return nil, fmt.Errorf("Could not find response with name '%s'", name)
	}
}

// Raw HTTP request accessor
func (r *requestData) HttpRequest() *http.Request {
	return r.httpRequest
}

//* Request params accessors */

// Param returns the value of a given request parameter
// Request parameters are defined in the action definitions of a resource definition. They appear in the action path
// with the mux capture syntax (e.g. "/users/{userId}" defines the "userId" parameter). They may also appear in the
// query string (e.g. the path "/users/{userId}?filter={filter}" defines two parameters: "userId" and "filter").
func (r *requestData) Param(name string) interface{} {
	return (*r.params)[name]
}

// ParamString returns the value of a given string request parameter, see Param() above.
func (r *requestData) ParamString(name string) string {
	res, _ := (*r.params)[name].(string)
	return res
}

// ParamInt returns the value of a given integer request parameter, see Param() above.
func (r *requestData) ParamInt(name string) int64 {
	res, _ := (*r.params)[name].(int64)
	return res
}

// ParamInt returns the value of a given boolean request parameter, see Param() above.
func (r *requestData) ParamBool(name string) bool {
	res, _ := (*r.params)[name].(bool)
	return res
}

// ParamInt returns the value of a given float parameter, see Param() above.
func (r *requestData) ParamFloat(name string) float64 {
	res, _ := (*r.params)[name].(float64)
	return res
}

// ParamInt returns the value of a given time parameter, see Param() above.
func (r *requestData) ParamTime(name string) time.Time {
	res, _ := (*r.params)[name].(time.Time)
	return res
}

/* Request payload accessor */

// Payload returns a pointer to the request payload as a struct whose type is defined by the model blueprint
func (r *requestData) Payload() interface{} {
	return r.payload
}

// RawRequest returns the underlying http request
func (r *requestData) RawRequest() *http.Request {
	return r.httpRequest
}

// A request handler implements the standard http HandlerFunc method for a single controller action.
type requestHandler struct {
	action       *Action       // Definition of action
	controller   Controller    // Instance of controller
	actionMethod reflect.Value // Action method to be invoked
	actionName   string        // Action name
}

// Factory method
func newRequestHandler(actionName string, action *Action, controller Controller) (*requestHandler, error) {
	actionMethod := reflect.ValueOf(controller).MethodByName(actionName)
	if actionMethod == (reflect.Value{}) {
		return nil, fmt.Errorf("No method '%s' in controller %v", actionName, reflect.TypeOf(controller))
	} else {
		return &requestHandler{action, controller, actionMethod, actionName}, nil
	}
}

// ServeHTTP implements the standard net/http HandlerFunc function.
// The steps involved here are:
//   1. Parse and validate request parameters if any
//   2. Parse and validate request payload (a.k.a. body) if any
//   3. Call controller method with resulting request struct
func (handler *requestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	params, err := handler.loadParams(req)
	if err != nil {
		handler.respondError(w, 400, "InvalidParam", err)
		return
	}
	payload, err := handler.loadPayload(req)
	if err != nil {
		handler.respondError(w, 400, "InvalidPayload: ", err)
		return
	}
	request := &requestData{&handler.action.pResponses, payload, params, req, nil}
	handler.actionMethod.Call([]reflect.Value{reflect.ValueOf(request)})
	res := request.response
	if res == nil {
		handler.respondError(w, 500, "InvalidResponse", errors.New("Response not initialized"))
		return
	}
	if err := handler.action.ValidateResponse(res); err != nil {
		handler.respondError(w, 500, "InvalidResponse: ", err)
		return
	}
	header := w.Header()
	for name, value := range *res.Header() {
		header[name] = value
	}
	w.WriteHeader(res.Status())
	if body, err := json.Marshal(res.Body()); err != nil {
		handler.respondError(w, 500, "Failed to serialize body", err)
	} else {
		w.Write([]byte(body))
	}
	parts := res.Parts()
	if len(parts) > 0 {
		m := multipart.NewWriter(w)
		for _, part := range parts {
			if body, err := json.Marshal(part.Body()); err != nil {
				handler.respondError(w, 500, "Failed to serialize part body", err)
			} else if err := m.WriteField(part.PartId(), string(body)); err != nil {
				handler.respondError(w, 500, "Failed to write part "+part.PartId(), err)
			}
		}
	}
}

// respondError writes back an error response using the given status, title (error summary) and error.
func (handler *requestHandler) respondError(w http.ResponseWriter, status int, title string, err error) {
	body := fmt.Sprintf("%s: %s\r\n", title, err.Error())
	w.WriteHeader(status)
	w.Write([]byte(body))
}

// loadParams loads the values from the request url and applies the validation rules defined in the action definition
// Parameters are defined in the action definition path and query string.
func (handler *requestHandler) loadParams(request *http.Request) (*map[string]interface{}, error) {
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
	return &params, nil
}

// loadPayload loads the payload attribute values from the request body and apply the validation rules defined in the
// payload attributes of the action definition.
// This function supports loading form encoding, multi-part form encoding and JSON encoding bodies.
// The result is then loaded into an instance of the action payload blueprint.
func (handler *requestHandler) loadPayload(request *http.Request) (interface{}, error) {
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
	payload, err := action.pPayload.Load(parsed)
	if err != nil {
		return nil, fmt.Errorf("Failed to load request payload: %s", err.Error())
	}

	return payload, nil
}

// loadValue loads a single value given a name, an incoming value and an attribute definition.
// This method both coerces the value if needed and validates it against the attribute definition.
// It returns the coerced value on success or an error on failure.
func (handler *requestHandler) loadValue(name string, raw interface{}, attr *Attribute) (interface{}, error) {
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
