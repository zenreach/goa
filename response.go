package goa

import "fmt"
import "regexp"
import "mime"
import "strings"
import "net/http"

// Response header definitions, map of definitions keyed by header name. There are two kinds of definitions:
//   - Regexp definitions consist of strings starting and ending with slash ("/").
//   - Exact matches consist of strings that do not start or do not end with slash (or neither).
// All action responses are validated against provided header defintions, at least one of the response headers must
// match each definition.
type Headers map[string]string

// Response definitions dictate the set of valid responses a given action may return.
// A response definition describes the response status code, media type and compulsory headers.
// The 'Location' header is called out as it is a common header returned by actions that create resources
// A multipart response definition may also describe compulsory headers for its parts.
type Response struct {
	Description string    // Description used by documentation
	Status      int       // Response status code
	MediaType   MediaType // Response media type if any
	Location    string    // Response 'Location' header validation, enclose value in / for regexp behavior
	Headers     Headers   // Response header validations, enclose values in / for regexp behavior
	Parts       *Response // Response part definitions if any

	// Internal fields

	resource *Resource // Parent resource definition
}

// Factory method to create corresponding responses
func (d *Response) NewResponse() ResponseData {
	return &standardResponse{definition: d}
}

// Provide helper methods for creating HTTP response from status
type f int

// HTTP Response factory
var Http f

func (f) Continue() Response {
	return Response{Status: 100, Description: "100 Continue"}
}
func (f) Ok() Response {
	return Response{Status: 200, Description: "200 OK"}
}
func (f) Created() Response {
	return Response{Status: 201, Description: "201 Created"}
}
func (f) Accepted() Response {
	return Response{Status: 202, Description: "202 Accepted"}
}
func (f) NonAuthoritative() Response {
	return Response{Status: 202, Description: "203 Non-Authoritative Information"}
}
func (f) NoContent() Response {
	return Response{Status: 204, Description: "204 No Content"}
}
func (f) ResetContent() Response {
	return Response{Status: 205, Description: "205 Reset Content"}
}
func (f) PartialContent() Response {
	return Response{Status: 206, Description: "206 Partial Content"}
}
func (f) MultipleChoices() Response {
	return Response{Status: 300, Description: "300 Multiple Choices"}
}
func (f) MovedPermanently() Response {
	return Response{Status: 301, Description: "301 Moved Permanently"}
}
func (f) Found() Response {
	return Response{Status: 302, Description: "302 Found"}
}
func (f) SeeOther() Response {
	return Response{Status: 303, Description: "303 See Other"}
}
func (f) NotModified() Response {
	return Response{Status: 304, Description: "304 Not Modified"}
}
func (f) UseProxy() Response {
	return Response{Status: 305, Description: "305 Use Proxy"}
}
func (f) TemporaryRedirect() Response {
	return Response{Status: 307, Description: "307 Temporary Redirect"}
}
func (f) BadRequest() Response {
	return Response{Status: 400, Description: "400 Bad Request"}
}
func (f) Unauthorized() Response {
	return Response{Status: 401, Description: "401 Unauthorized"}
}
func (f) PaymentRequired() Response {
	return Response{Status: 402, Description: "402 Payment Required"}
}
func (f) Forbidden() Response {
	return Response{Status: 403, Description: "403 Forbidden"}
}
func (f) NotFound() Response {
	return Response{Status: 404, Description: "404 Not Found"}
}
func (f) MethodNotAllowed() Response {
	return Response{Status: 405, Description: "405 Method Not Allowed"}
}
func (f) NotAcceptable() Response {
	return Response{Status: 406, Description: "406 Not Acceptable"}
}
func (f) ProxyAuthRequired() Response {
	return Response{Status: 407, Description: "407 Proxy Authentication Required"}
}
func (f) RequestTimeout() Response {
	return Response{Status: 408, Description: "408 Request Timeout"}
}
func (f) Conflict() Response {
	return Response{Status: 409, Description: "409 Conflict"}
}
func (f) Gone() Response {
	return Response{Status: 410, Description: "410 Gone"}
}
func (f) LengthRequired() Response {
	return Response{Status: 411, Description: "411 Length Required"}
}
func (f) PreconditionFailed() Response {
	return Response{Status: 412, Description: "412 Precondition Failed"}
}
func (f) RequestEntityTooLarge() Response {
	return Response{Status: 413, Description: "413 Request Entity Too Large"}
}
func (f) RequestUriTooLong() Response {
	return Response{Status: 414, Description: "414 Request-URI Too Long"}
}
func (f) UnsupportedMediaType() Response {
	return Response{Status: 415, Description: "415 Unsupported Media Type"}
}
func (f) RequestRangeNotSatisfiable() Response {
	return Response{Status: 416, Description: "416 Request Range Not Satisfiable"}
}
func (f) ExpectationFailed() Response {
	return Response{Status: 417, Description: "417 Expectation Failed"}
}
func (f) InternalError() Response {
	return Response{Status: 500, Description: "500 Internal Error"}
}
func (f) NotImplemented() Response {
	return Response{Status: 501, Description: "501 Not Implemented"}
}
func (f) BadGateway() Response {
	return Response{Status: 502, Description: "502 Bad Gateway"}
}
func (f) ServiceUnavailable() Response {
	return Response{Status: 503, Description: "503 Service Unavailable"}
}
func (f) GatewayTimeout() Response {
	return Response{Status: 504, Description: "504 Gateway Timeout"}
}
func (f) HTTPVersionNotSupported() Response {
	return Response{Status: 505, Description: "505 HTTP Version Not Supported"}
}

// Validate checks response against definition.
// Returns error if validation fails, nil otherwise.
func (d *Response) Validate(r ResponseData) error {
	if d.Status == 500 {
		return nil // Already an error
	}
	if d.Status > 0 {
		if r.Status() != d.Status {
			return fmt.Errorf("Value of response status does not match response definition (value is '%v', definition's is '%v')", r.Status(), d.Status)
		}
	}
	header := r.Header()
	if len(d.Location) > 0 {
		val := header.Get("Location")
		if !d.matches(val, d.Location) {
			return fmt.Errorf("Value of response header Location does not match response definition (value is '%s', definition's is '%s')", val, d.Location)
		}
	}
	if len(d.Headers) > 0 {
		for name, value := range d.Headers {
			val := strings.Join((*header)[http.CanonicalHeaderKey(name)], ",")
			if !d.matches(val, value) {
				return fmt.Errorf("Value of response header %s does not match response definition (value is '%s', definition's is '%s')", name, val, value)
			}
		}
	}
	media_type := d.MediaType
	if (&media_type).IsEmpty() {
		media_type = d.resource.MediaType
	}
	id := media_type.Identifier
	if len(id) > 0 {
		parsed, _, err := mime.ParseMediaType(id)
		if err != nil {
			return fmt.Errorf("Invalid media type identifier '%s': %s", id, err.Error())
		}
		val := strings.Join((*header)["Content-Type"], ",")
		if parsed != strings.ToLower(val) {
			return fmt.Errorf("Value of response header Content-Type does not match response definition (value is '%s', definition's is '%s')", val, parsed)
		}
	}
	if d.Parts != nil {
		for name, part := range r.Parts() {
			if err := d.Parts.Validate(part); err != nil {
				msg := err.Error()
				msg = strings.ToLower(string(msg[0])) + msg[1:]
				return fmt.Errorf("Invalid response part %s, %s", name, msg)
			}
		}
	}
	return nil
}

// matches checks whether string value matches definition, returns true if it does, false otherwise
// If definition is a string that starts and ends with "/" then value is matched against a regexp built from definition
// otherwise value is compared directly with definition
func (d *Response) matches(value, match string) bool {
	ok := false
	matches := matchRegexp.FindStringSubmatch(match)
	if len(matches) > 0 {
		ok, _ = regexp.MatchString(value, matches[1])
	} else {
		ok = (value == match)
	}
	return ok
}

var matchRegexp = regexp.MustCompile("^/(.*)/$")
