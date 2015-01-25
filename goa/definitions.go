package main

import (
	"go/ast"
)

// Mediatype definition: defines identifier
type mediaTypeDef struct {
	identifier string
	spec       *ast.TypeSpec
}

// Resource directives: version and default media type
// Interface that defines resource actions
type resourceDef struct {
	name       string
	apiVersion string
	mediaType  string
	basePath   string
	actions    map[string]*actionDef
	spec       *ast.TypeSpec
}

// Resource action directives: route and responses
type actionDef struct {
	name      string
	method    string
	path      string
	responses map[int]*responseDef
	field     *ast.Field
}

// Response directives: body and headers
type responseDef struct {
	code      int
	mediaType string
	headers   map[string]string
}

// Controller directive: specifies resource being implemented
type controllerDef struct {
	resource string
	spec     *ast.TypeSpec
}
