package main

import (
	"go/doc"
)

// Mediatype definition: defines identifier
type mediaTypeDef struct {
	name       string    // go struct name
	identifier string    // Media type identifier
	docs       *doc.Type // Documentation
}

// Resource directives: version and default media type
// Interface that defines resource actions
type resourceDef struct {
	name       string                // go interface name
	apiVersion string                // API version - can be empty
	mediaType  string                // Media type identifier
	basePath   string                // Base path for all actions - can be empty
	actions    map[string]*actionDef // Resource action definitions
	docs       *doc.Type             // Documentation
}

// Resource action directives: route and responses
type actionDef struct {
	name      string               // Action name (method name)
	method    string               // Action HTTP method ("GET", "POST", etc.)
	path      string               // Action path relative to resource base path
	responses map[int]*responseDef // Response definitions
	docs      *doc.Func            // Documentation
}

// Response directives: body and headers
type responseDef struct {
	code      int               // HTTP status code
	mediaType string            // Media type identifier
	headers   map[string]string // HTTP headers
}

// Controller directive: specifies resource being implemented
type controllerDef struct {
	name     string    // go struct name
	resource string    // Resource interface name
	docs     *doc.Type // Documentation
}
