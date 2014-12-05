package goa

import "crypto/rand"
import "net/http"

// ResponseData provides access to the HTTP response data.
// goa provides a default implementation and various factory methods for building the response.
// Actions may alternatively initialize the Response field of the request object they receive as argument with their
// own implementation of the interface.
type ResponseData interface {
	Status() int                    // HTTP response status
	Header() *http.Header           // HTTP response headers
	Body() interface{}              // HTTP response body
	Parts() map[string]ResponseData // Multipart response parts if any
	PartId() string                 // Multipart response inner part id if any
}

// The ResponseBuilder interface exposes methods use by actions to initialize the response.
type ResponseBuilder interface {
	SetHeader(name, value string)
	AddHeader(name, value string)
	SetBody(body string)
	AddPart(part ResponseData)
	Response() ResponseData
}

// Default Response implementation
type standardResponse struct {
	definition *Response
	status     int
	header     *http.Header
	body       interface{}
	partId     string
	parts      map[string]ResponseData
}

/* Methods used by controllers to initialize response */

// Response returns a ResponseData interface implementation
// Standard response object implements both ResponseData and ResponseBuilder interfaces
func (r *standardResponse) Response() ResponseData {
	return r
}

// SetHeader sets response header
func (r *standardResponse) SetHeader(name, value string) {
	r.header.Set(name, value)
}

// AddHeader adds response header (appends to any existing value associated with name)
func (r *standardResponse) AddHeader(name, value string) {
	r.header.Add(name, value)
}

// Set response body
func (r *standardResponse) SetBody(body string) {
	r.body = body
}

// AddPart adds part to multipart response
// A part contains the same elements as a standard response (headers, body etc.)
func (r *standardResponse) AddPart(part ResponseData) {
	r.parts[part.PartId()] = part
}

/* ResponseData interface implementation */

func (r *standardResponse) Status() int {
	if r.status > 0 {
		return r.status
	}
	return 200
}

func (r *standardResponse) Header() *http.Header {
	return r.header
}

func (r *standardResponse) Body() interface{} {
	return r.body
}

func (r *standardResponse) Parts() map[string]ResponseData {
	return r.parts
}

func (r *standardResponse) PartId() string {
	if len(r.partId) == 0 {
		r.partId = randStr(20)
	}
	return r.partId
}

/* Helper method used to generate random strings */
/* Assuming perfect distribution of randomizer, chances of getting two identical values is 1 over 62^size */

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func randStr(size int) string {
	var bytes = make([]byte, size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
