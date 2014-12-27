package goa

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
)

// A controller in go can have any type
type Controller interface{}

// The ResponseBuilder interface exposes methods use by actions to initialize
// the HTTP response. This interface is implemented by the Request struct and by
// structs returned by its `AddPart()` method.
// All methods return a ResponseBuilder interface so they can be chained.
//
// Examples:
//     r.Respond("").WithStatus(204)
//
//     // json contains a string resuling from JSON encoding
//     r.Respond(json)
//      .WithStatus(200)
//      .WithHeader("Content-Type", "application/json")
//
//     // Multipart (bulk) creation response
//     // `resources` is the collection of resources that got created
//     for _, resource := range resources {
//         part := r.AddPart(resource.id)
//         part.Respond("")
//             .WithStatus(201)
//             .WithHeader("Location", resource.href)
//     }
//     r.Respond("{created:"+strconv.Itoa(len(resources))+"}").WithStatus(200)
type ResponseBuilder interface {
	// Set response body (empty by default)
	Respond(body string) ResponseBuilder
	// Set response status code (200 by default)
	WithStatus(status int) ResponseBuilder
	// Set a response header
	WithHeader(name, value string) ResponseBuilder
	// Add a multipart response part
	AddPart(partId string) ResponseBuilder
}

// A goa request includes all the information needed by the controller action
// to perform. It also implements ResponseBuilder so that actions may use that
// object to build the action response.
// All controller actions take a pointer to a Request struct as first argument.
type Request struct {
	// Underlying HTTP request
	Raw *http.Request
	// Parsed parameters
	Params map[string]interface{}
	// Parsed payload
	Payload interface{}
	// Underlying HTTP response writer
	ResponseWriter http.ResponseWriter

	// Request response built through RequestBuilder interface
	response *standardResponse
}

// Respond sets the response body
func (r *Request) Respond(body string) ResponseBuilder {
	r.response.body = body
	return r
}

// WithStatus sets the current response status
func (r *Request) WithStatus(status int) ResponseBuilder {
	r.response.status = status
	return r
}

// WithHeader sets a header on the current response
// It returns the controller so that it can be chained with other
// response builder methods.
func (r *Request) WithHeader(name, value string) ResponseBuilder {
	r.response.header.Set(name, value)
	return r
}

// AddPart returns a multipart response part
// The part should be initialized using the ResponseBuilder methods
func (r *Request) AddPart(partId string) ResponseBuilder {
	r.response.parts[partId] = new(standardResponse)
	return r
}

// Default Response implementation
type standardResponse struct {
	status int
	header http.Header
	body   string
	partId string
	parts  map[string]*standardResponse
}

// Status is a simple method used to access the response status.
// Defaults status to 200.
func (r *standardResponse) Status() int {
	if r.status != 0 {
		return r.status
	} else {
		return 200
	}
}

// sendResponse sends the response if GetResponseWriter has not been called,
// does nothing otherwise.
func (r *Request) sendResponse(action *Action) {
	res := r.response
	if err := action.ValidateResponse(res); err != nil {
		r.respondError(500, "InvalidResponse", err)
		return
	}
	w := r.ResponseWriter
	w.WriteHeader(res.Status())
	header := w.Header()
	for name, value := range res.header {
		header[name] = value
	}
	w.Write([]byte(res.body))
	parts := res.parts
	if len(parts) > 0 {
		m := multipart.NewWriter(w)
		for id, part := range parts {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", part.Status(), http.StatusText(part.Status())))
			for name, value := range part.header {
				buffer.WriteString(fmt.Sprintf("%s: %s\r\n", name, value))
			}
			buffer.WriteString("\r\n")
			buffer.WriteString(part.body)
			if err := m.WriteField(id, buffer.String()); err != nil {
				r.respondError(500, "Failed to write part "+id, err)
				return
			}
		}
	}
}

// respondError writes back an error response using the given status, title
// (error summary) and error.
func (r *Request) respondError(status int, title string, err error) {
	body := fmt.Sprintf("%s: %s\r\n", title, err.Error())
	r.ResponseWriter.WriteHeader(status)
	r.ResponseWriter.Write([]byte(body))
}
