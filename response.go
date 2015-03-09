package goa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// An action response definition
// Specifies a status, a body and header values.
// Usage:
//     responseContent := ...
//     r := goa.Http.Ok().WithBody(responseContent)
type Response struct {
	Status int       // Response status code
	Body   io.Reader // Response body reader
	Header http.Header
}

// Sets the body of the response with given parameter. Actual behavior depends on parameter type:
// if the parameter is a string then it is used as is otherwise it is json encoded.
// Calling this method with `nil` does nothing.
// WithBody returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithBody(body interface{}) *Response {
	if body == nil {
		return r
	}
	if b, ok := body.(string); ok {
		r.Body = strings.NewReader(b)
		return r
	}
	if b, err := json.Marshal(body); err != nil {
		r.Body = strings.NewReader(fmt.Sprintf("API Bug: failed to serialize response: %s", err.Error()))
		r.Status = 500
	} else {
		r.Body = bytes.NewBuffer(b)
	}
	return r
}

// WithLocation sets the response Location header.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithLocation(l string) *Response {
	return r.WithHeader("Location", l)
}

// WithHeader sets the given header with the given value.
// It returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithHeader(name string, value string) *Response {
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	r.Header.Set(name, value)
	return r
}
