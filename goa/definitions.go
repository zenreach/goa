package main

import (
	"io"
)

// A parsed goa controller, resource or media type definition
type definition interface {
	// Generate code for definition
	generate(io.Writer, *report)
}

// Mediatype definition: defines identifier
type mediaTypeDef struct {
	identifier string
	spec       *ast.TypeSpec
}

// Resource directives: version and default media type
// Interface that defines resource actions
type resourceDef struct {
	apiVersion string
	mediaType  string
	actions    map[string]*ActionDef
	spec       *ast.TypeSpec
}

// Resource action directives: route and responses
type ActionDef struct {
	route     string
	responses map[int]*ResponseDef
}

// Response directives: body and headers
type ResponseDef struct {
	mediaType string
	headers   map[string]string
}

// Controller directive: specifies resource being implemented
type controllerDef struct {
	resource string
	spec     *ast.TypeSpec
}

type report struct {
	
}

func (m *mediaTypeDef) generate(output io.Writer) {
}

func (m *resourceDef) generate(output io.Writer) {
}

func (m *controllerDef) generate(output io.Writer) {
}
