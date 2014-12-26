package goa

import "fmt"
import "regexp"
import "mime"
import "strings"
import "net/http"

// Response header definitions, map of definitions keyed by header name. There
// are two kinds of definitions:
//   - Regexp definitions consist of strings starting and ending with slash "/".
//   - Exact matches consist of strings that do not start or do not end with
//     slash (or neither).
// All action responses are validated against provided header defintions, at
// least one of the response headers must match each definition.
type Headers map[string]string

// Response definitions dictate the set of valid responses a given action may
// return. A response definition describes the response status code, media type
// and compulsory headers.
// The 'Location' header is called out as it is a common header returned by
// actions that create resources. A multipart response definition may also
// describe compulsory headers for its parts.
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

// Provide helper methods for creating HTTP response from status
type f int

// HTTP Response factory
var Http f

func vanillaResponse(status int) Response {
	return Response{Status: status, Description: http.StatusText(status)}
}

func (f) Continue() Response                   { return vanillaResponse(100) }
func (f) Ok() Response                         { return vanillaResponse(200) }
func (f) Created() Response                    { return vanillaResponse(201) }
func (f) Accepted() Response                   { return vanillaResponse(202) }
func (f) NonAuthoritative() Response           { return vanillaResponse(203) }
func (f) NoContent() Response                  { return vanillaResponse(204) }
func (f) ResetContent() Response               { return vanillaResponse(205) }
func (f) PartialContent() Response             { return vanillaResponse(206) }
func (f) MultipleChoices() Response            { return vanillaResponse(300) }
func (f) MovedPermanently() Response           { return vanillaResponse(301) }
func (f) Found() Response                      { return vanillaResponse(302) }
func (f) SeeOther() Response                   { return vanillaResponse(303) }
func (f) NotModified() Response                { return vanillaResponse(304) }
func (f) UseProxy() Response                   { return vanillaResponse(305) }
func (f) TemporaryRedirect() Response          { return vanillaResponse(307) }
func (f) BadRequest() Response                 { return vanillaResponse(400) }
func (f) Unauthorized() Response               { return vanillaResponse(401) }
func (f) PaymentRequired() Response            { return vanillaResponse(402) }
func (f) Forbidden() Response                  { return vanillaResponse(403) }
func (f) NotFound() Response                   { return vanillaResponse(404) }
func (f) MethodNotAllowed() Response           { return vanillaResponse(405) }
func (f) NotAcceptable() Response              { return vanillaResponse(406) }
func (f) ProxyAuthRequired() Response          { return vanillaResponse(407) }
func (f) RequestTimeout() Response             { return vanillaResponse(408) }
func (f) Conflict() Response                   { return vanillaResponse(409) }
func (f) Gone() Response                       { return vanillaResponse(410) }
func (f) LengthRequired() Response             { return vanillaResponse(411) }
func (f) PreconditionFailed() Response         { return vanillaResponse(412) }
func (f) RequestEntityTooLarge() Response      { return vanillaResponse(413) }
func (f) RequestUriTooLong() Response          { return vanillaResponse(414) }
func (f) UnsupportedMediaType() Response       { return vanillaResponse(415) }
func (f) RequestRangeNotSatisfiable() Response { return vanillaResponse(416) }
func (f) ExpectationFailed() Response          { return vanillaResponse(417) }
func (f) InternalError() Response              { return vanillaResponse(500) }
func (f) NotImplemented() Response             { return vanillaResponse(501) }
func (f) BadGateway() Response                 { return vanillaResponse(502) }
func (f) ServiceUnavailable() Response         { return vanillaResponse(503) }
func (f) GatewayTimeout() Response             { return vanillaResponse(504) }
func (f) HTTPVersionNotSupported() Response    { return vanillaResponse(505) }

// Validate checks response against definition.
// Returns error if validation fails, nil otherwise.
func (d *Response) Validate(r *standardResponse) error {
	if d.Status == 500 {
		return nil // Already an error, protect against infinite loops
	}
	if d.Status > 0 {
		if r.Status() != d.Status {
			return fmt.Errorf("Value of response status does not match response definition (value is '%v', definition's is '%v')", r.Status(), d.Status)
		}
	}
	header := r.header
	if len(d.Location) > 0 {
		val := header.Get("Location")
		if !d.matches(val, d.Location) {
			return fmt.Errorf("Value of response header Location does not match response definition (value is '%s', definition's is '%s')", val, d.Location)
		}
	}
	if len(d.Headers) > 0 {
		for name, value := range d.Headers {
			val := strings.Join(header[http.CanonicalHeaderKey(name)], ",")
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
		val := strings.Join(header["Content-Type"], ",")
		if parsed != strings.ToLower(val) {
			return fmt.Errorf("Value of response header Content-Type does not match response definition (value is '%s', definition's is '%s')", val, parsed)
		}
	}
	if d.Parts != nil {
		for name, part := range r.parts {
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
