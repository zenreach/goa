package goa

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

var (
	handlerProviders map[string]HandlerProvider
)

// Request handler.
type Handler struct {
	// Underlying http response writer
	W http.ResponseWriter
	// Underlying http request
	R *http.Request
}

// NewHandler instantiates a new request handler for an action on the resource with given name.
// Handlers are registered when mounted under the app.
// This function is called by the bootstrapped code.
func NewHandler(resName string, w http.ResponseWriter, r *http.Request) (*Handler, error) {
	provider, ok := handlerProviders[resName]
	if !ok {
		return nil, fmt.Errorf("No handler associated with %s", resName)
	}
	return provider(r, w), nil
}

// LoadRequestBody decodes the request body. It returns the decoded content or an array of decoded
// contents in the case of a multipart body.
// The following content types are  supported:
// application/json, text/json, <anything>+json: body is decoded with the JSON decoder.
// application/x-www-form-urlencoded: body is read as a url encoded form.
// multipart/<anything>: each part is decoded using the decoder returned by applying this same
// algorithm to the part content-type header.
// Returns an error if the content type is not supported or decoding fails.
func (h *Handler) LoadRequestBody(r *http.Request) (interface{}, error) {
	mediaType, params, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("invalid request media type: %s", err)
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(request.Body, params["boundary"])
		var contents []interface{}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return contents, nil
			}
			if err != nil {
				return nil, fmt.Errorf("fail to read part enveloppe: %s", err)
			}
			c, err := h.loadSingleBody(p.Header.Get("Content-Type"), p)
			if err != nil {
				return nil, fmt.Errorf("fail to decode part body: %s", err)
			}
			contents = append(contents, c)
		}
	}
	return h.loadSingleBody(mediaType, r.Body)
}

// InitStruct loads data from a map into a struct recursively.
func (h *Handler) InitStruct(inited interface{}, data map[string]interface{}) error {
	initType := reflect.TypeOf(inited)
	if initType == nil || initType.Kind() != reflect.Ptr {
		return fmt.Errorf("invalid inited value, must be a pointer - got %v", initType)
	}
	sType := initType.Elem()
	if sType == nil || sType.Kind() != reflect.Struct {
		return fmt.Errorf("invalid inited value, must be a pointer on struct - got pointer on %v", sType)
	}
	value := reflect.Zero(sType)
	if err := h.initData(reflect.ValueOf(value), reflect.ValueOf(data), ""); err != nil {
		return err
	}
	reflect.ValueOf(inited).Elem().Set(value)
	return nil
}

// WriteResponse writes the given HTTP response using the handler responser writer.
func (h *Handler) WriteResponse(r *Response) {
	var b []byte
	if len(r.Body) > 0 {
		var err error
		if b, err = json.Marshal(r.Body); err != nil {
			RespondInternalError(fmt.Errorf("API bug, failed to serialize response body: %s", err))
			return
		}
	}
	w := h.W
	if len(r.Headers) > 0 {
		h := w.Header()
		for n, v := range r.Headers {
			h.Set(n, v)
		}
	}
	w.WriteHeader(r.Status)
	w.Write(b)
}

// loadSingleBody is a helper function used by LoadRequestBody to decode the content of a single
// HTTP request body encoded using the media type identified by mt. See LoadRequestBody for more
// details.
func (h *Handler) loadSingleBody(mt string, body io.Reader) (interface{}, error) {
	slurp, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("fail to read body: %s", err)
	}
	if strings.Contains(mt, "form-urlencoded") {
		// The code below is from http://golang.org/src/net/http/request.go?s=23467:23502#L769
		// Is there a better way?
		maxFormSize := int64(1<<63 - 1)
		if _, ok := body.(*maxBytesReader); !ok {
			maxFormSize = int64(10 << 20) // 10 MB is a lot of text.
			reader = io.LimitReader(body, maxFormSize+1)
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("fail to read form body: %s", err)
		}
		if int64(len(b)) > maxFormSize {
			return nil, fmt.Errorf("request body too large")
		}
		vs, err = url.ParseQuery(string(b))
		if err != nil {
			return nil, fmt.Errorf("fail to decode form body: %s", err)
		}
		values := make(map[string]interface{})
		for n, v := range vs {
			values[n] = v
		}
		return values, nil
	} else if strings.HasSuffix(mt, "json") {
		decoder := json.NewDecoder(body)
		var decoded interface{}
		err := decoder.Decode(&decoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %s", err)
		}
		return decoded, nil
	} else {
		return nil, fmt.Errorf("unsupported content type '%s'", mt)
	}
}

// Initialize data structure recursively using provided data (map of string to interface).
// Last argument is path to field currently being init'ed (using dot notation).
func (h *Handler) initData(value reflect.Value, data reflect.Value, attPrefix string) error {
	for _, k := range data.MapKeys() {
		key := k.String()
		if len(attPrefix) > 0 {
			key = attPrefix + "." + key
		}
		fieldName, _ := b.fieldByProp[key]
		f := value.FieldByName(fieldName)
		if !f.IsValid() {
			return fmt.Errorf("unknown %v field '%s'", value.Type(), fieldName)
		}
		if !f.CanSet() {
			return fmt.Errorf("%v field '%s' cannot be written to, is it public?",
				value.Type(), fieldName)
		}
		val := data.MapIndex(k).Elem()
		if val.Type().Kind() == reflect.Map {
			if err := h.initData(f, val, key); err != nil {
				return err
			}
		} else {
			if err := h.setFieldValue(f, val, fieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

// setFieldValue loads given value into given struct field.
// Value type must be a JSON schema primitive type.
func (h *Handler) setFieldValue(field, value reflect.Value, fieldName string) error {
	if err := b.validateFieldKind(field, value.Kind(), fieldName); err != nil {
		return fmt.Errorf("field '%s': %s", fieldName, err)
	}
	// value must be a string, int, float64, bool, array or map of values
	switch value.Kind() {
	case reflect.String:
		field.SetString(value.String())
	case reflect.Int:
		i := value.Int()
		if !field.OverflowInt(i) {
			field.SetInt(i)
		} else {
			return fmt.Errorf("field '%s': integer value too big", fieldName)
		}
	case reflect.Float64:
		f := value.Float()
		if !field.OverflowFloat(f) {
			field.SetFloat(f)
		} else {
			return fmt.Errorf("field '%s': float value too big", fieldName)
		}
	case reflect.Bool:
		field.SetBool(value.Bool())
	case reflect.Array:
		field.Set(reflect.MakeSlice(value.Elem().Type(), value.Len(), value.Len()))
		for i := 0; i < value.Len(); i++ {
			if err := h.setFieldValue(field.Index(i), value.Index(i),
				fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
				return fmt.Errorf("field '%s' item %d: %s", fieldName, i, err)
			}
		}
	}

	return nil
}

// Helper function used to validate kind of struct field value against attribute type
func (h *Handler) validateFieldKind(field reflect.Value, kind reflect.Kind, name string) error {
	if field.Kind() != kind {
		return fmt.Errorf("invalid value type '%v'", kind)
	}
	return nil
}
