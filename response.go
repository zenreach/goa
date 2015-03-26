package goa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// An action response
// Specifies a status, a body and header values.
// Usage:
//     responseContent := ...
//     r := goa.Ok().WithBody(responseContent)
//     r.Write(w)
type Response struct {
	Status int       // Response status code
	Body   io.Reader // Response body reader
	Header http.Header
}

// Response factory methods
// Example usage:
//     r := goa.Created().WithLocation(href)

func Continue() *Response                   { return vanillaResponse(100) }
func Ok() *Response                         { return vanillaResponse(200) }
func Created() *Response                    { return vanillaResponse(201) }
func Accepted() *Response                   { return vanillaResponse(202) }
func NonAuthoritative() *Response           { return vanillaResponse(203) }
func NoContent() *Response                  { return vanillaResponse(204) }
func ResetContent() *Response               { return vanillaResponse(205) }
func PartialContent() *Response             { return vanillaResponse(206) }
func MultipleChoices() *Response            { return vanillaResponse(300) }
func MovedPermanently() *Response           { return vanillaResponse(301) }
func Found() *Response                      { return vanillaResponse(302) }
func SeeOther() *Response                   { return vanillaResponse(303) }
func NotModified() *Response                { return vanillaResponse(304) }
func UseProxy() *Response                   { return vanillaResponse(305) }
func TemporaryRedirect() *Response          { return vanillaResponse(307) }
func BadRequest() *Response                 { return vanillaResponse(400) }
func Unauthorized() *Response               { return vanillaResponse(401) }
func PaymentRequired() *Response            { return vanillaResponse(402) }
func Forbidden() *Response                  { return vanillaResponse(403) }
func NotFound() *Response                   { return vanillaResponse(404) }
func MethodNotAllowed() *Response           { return vanillaResponse(405) }
func NotAcceptable() *Response              { return vanillaResponse(406) }
func ProxyAuthRequired() *Response          { return vanillaResponse(407) }
func RequestTimeout() *Response             { return vanillaResponse(408) }
func Conflict() *Response                   { return vanillaResponse(409) }
func Gone() *Response                       { return vanillaResponse(410) }
func LengthRequired() *Response             { return vanillaResponse(411) }
func PreconditionFailed() *Response         { return vanillaResponse(412) }
func RequestEntityTooLarge() *Response      { return vanillaResponse(413) }
func RequestUriTooLong() *Response          { return vanillaResponse(414) }
func UnsupportedMediaType() *Response       { return vanillaResponse(415) }
func RequestRangeNotSatisfiable() *Response { return vanillaResponse(416) }
func ExpectationFailed() *Response          { return vanillaResponse(417) }
func InternalError() *Response              { return vanillaResponse(500) }
func NotImplemented() *Response             { return vanillaResponse(501) }
func BadGateway() *Response                 { return vanillaResponse(502) }
func ServiceUnavailable() *Response         { return vanillaResponse(503) }
func GatewayTimeout() *Response             { return vanillaResponse(504) }
func HTTPVersionNotSupported() *Response    { return vanillaResponse(505) }

// WithBody initializes the body of the response.
// The actual behavior depends on the type of body: if body is a string or an io.Reader then it is
// stored as is otherwise it is first json encoded.
// Calling this method with nil does nothing.
// WithBody returns the response so it can be chained with other WithXXX methods.
func (r *Response) WithBody(body interface{}) *Response {
	if body == nil {
		return r
	}
	switch b := body.(type) {
	case error:
		r.Body = strings.NewReader(b.Error())
	case string:
		r.Body = strings.NewReader(b)
	case io.Reader:
		r.Body = b
	default:
		if b, err := json.Marshal(body); err != nil {
			r.Body = strings.NewReader(fmt.Sprintf("API Bug: failed to serialize response: %s", err))
			r.Status = 500
		} else {
			r.Body = bytes.NewBuffer(b)
		}
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

// Write serializes the response body with JSON and writes it to the given response writer.
func (r *Response) Write(w http.ResponseWriter) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RespondInternalError(w, "API bug, failed to read response body: %s", err)
		return
	}
	r.Header.Set("Content-Length", strconv.Itoa(len(body)))
	writeHeaders(w, r)
	w.Write(body)
}

// Max number of bytes read and sent in each chunk when streaming response
const maxStreamChunkSizeBytes = 4096

// Stream uses chunk encoding to send blocks of data read from the response reader.
func (r *Response) Stream(w http.ResponseWriter) {
	writeHeaders(w, r)
	for {
		buffer := make([]byte, maxStreamChunkSizeBytes)
		read, err := r.Body.Read(buffer)
		if read > 0 {
			w.Write(buffer)
		}
		if err != nil {
			if err != io.EOF {
				w.Write([]byte(err.Error()))
			}
			break
		}
	}
}

// Write response headers and status code
func writeHeaders(w http.ResponseWriter, r *Response) {
	header := w.Header()
	for n, v := range r.Header {
		header[n] = v
	}
	w.WriteHeader(r.Status)
}

// vanillaResponse returns a default response for the given HTTP status code
func vanillaResponse(status int) *Response {
	return &Response{
		Status: status,
		Body:   strings.NewReader(http.StatusText(status)),
		Header: make(http.Header),
	}
}
