package goa

import (
	"net/http"
)

type Resource struct {
	Name        string
	Description string
	ApiVersion  string
	FullPath    string
   	Controller  Controller
	Actions     map[string]Action
}

type Action struct {
   	Name        string
	Description string
	Multipart   int
	Payload     *Model     // non-nil if action accepts a payload
	Params      Params
	Filters     Filters
	Views       []string
	Resource    *Resource  // Parent resource definition
	Routes      []Route    // Base URI to action including app base path and resource route prefix
	Responses   []Response // Action response definitions
}

type Response struct {
	Resource  *Resource         // Parent resource definition
	MediaType *MediaType        // Validated media type
	Parts     *Response // Response part definitions if any
}

// A route is the full url to an action request and its associated HTTP verb.
type Route struct {
	Verb             string // One of "GET", "POST", "PUT", "DELETE" etc.
	Path             string
	CapturePositions map[string]int
}