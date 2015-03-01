package goa

import (
	"fmt"
	"net/http"
	"net/url"
)

// Convenience struct used by generated code
type Hash map[string]interface{}

// A controller exposes methods used by actions to handle requests.
// The original net/http package response writer interface and http.Request
// struct are accessible via the `W` and `R` fields respectively.
type Controller struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Controller) Respond(code int) *Response {
	return &Response{Status: code}
}

func (c *Controller) RespondBadRequest(msg string) *Response {
	return &Response{}
}

func (c *Controller) RespondInternalError(msg string) *Response {
	return &Response{}
}

// CoerceParameter casts or parses request parameter into corresponding action
// method argument go type.
func CoerceParameter(param, name string, action *ActionDefinition) error {
	return nil
}

// ValidateParameter checks whether given request parameter validates the
// corresponding action definition parameter json schema.
func ValidateParameter(param, name string, action *ActionDefinition) error {
	return nil
}

// Check whether request has header or parameter version that matches 'version'
func CheckVersion(r *http.Request, version string) error {
	if len(version) == 0 {
		return nil
	}
	if r.Header.Get("X-Api-Version") == version {
		return nil
	}
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		if v, ok := params["api_version"]; ok {
			if v == version {
				return nil
			}
		}
	}
	return fmt.Errorf("Bad or missing API version. Specify with "+
		"\"?api_version=%s\" param or \"X-Api-Version=%s\" header.",
		version, version)
}
