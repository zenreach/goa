package goa

import (
	"fmt"
)

// Convenience struct used by generated code
type Hash map[string]interface{}

type Controller struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Controller) Respond(code int) *Response {
	return &Response{}
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
		"\"?api_version=%s\" param or \"X-API-VERSION=%s\" header.",
		version, version)
}
