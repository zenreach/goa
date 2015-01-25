package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"sort"
)

// Generator entry point: generate code for API
func generateApi(api *apiDescription, w io.Writer) errors {
	errs := errors{}
	names := make([]string, len(api.resources))
	idx := 0
	for name, _ := range api.resources {
		names[idx] = name
		idx += 1
	}
	sort.Strings(names)
	identifiers := make(map[string]bool, len(api.mediaTypes))
	for i, _ := range api.mediaTypes {
		identifiers[i] = true
	}
	for _, name := range names {
		resource, _ := api.resources[name]
		generateResource(resource, api, w)
		controller, err := api.resourceCompiler(resource)
		if err != nil {
			errs.add(err)
		} else {
			generateController(controller, api, w)
		}
		delete(identifiers, resource.mediaType)
		errs.addIf(generateMediaType(resource.mediaType, api, w))
	}
	for i, _ := range identifiers {
		errs.addIf(generateMediaType(i, api, w))
	}
	return errs
}

func generateMediaType(name string, api *apiDescription, o io.Writer) error {
	return nil
}

func generateResource(r *resourceDef, api *apiDescription, o io.Writer) error {
	schema, err := structToSchema(api.mediaTypes[r.mediaType].spec)
	if err != nil {
		return err
	}
	w := newWriter()
	w.w("//== %s ==\n\n", r.name)
	w.w("type goa_%s ResourceDefinition\n\n", r.name)
	w.w("var goa_%sSchema = %s\n\n", r.mediaType, schemaToSource(schema))
	w.w("func goa_Mount%sHandlers(app goa.Application) {\n", r.name)

	w.flush(o)
	return nil
}

func generateController(c *controllerDef, api *apiDescription, o io.Writer) error {
	return nil
}

// Generate JSON schema from arbitrary data structure.
// Struct field tags may be used to specify validation rules.
func generateJsonSchema(st *ast.StructType) (map[string]interface{}, error) {
	fields := st.Fields.List
	for _, field := range fields {
		typ := field.Type
		fmt.Printf("Type: %v+\n", typ)
		for _, name := range field.Names {
			fmt.Printf("Name: %v+\n", name)
		}
	}
	return map[string]interface{}{}, nil
}

// Json schema defining single data type
func typeSchema(t string) map[string]interface{} {
	return map[string]interface{}{"type": t}
}

func structToSchema(s *ast.TypeSpec) (interface{}, error) {
	return nil, nil
}

func schemaToSource(s interface{}) string {
	return ""
}

// Convenience wrapper around buffer
type writer struct {
	*bytes.Buffer
}

// Writer factory
func newWriter() writer {
	return writer{bytes.NewBuffer(make([]byte, 1024))}
}

// Writer write
func (b writer) w(text string, args ...interface{}) {
	b.WriteString(fmt.Sprintf(text, args...))
}

// Flush writer buffer into given io
func (b writer) flush(output io.Writer) {
	output.Write(b.Bytes())
}
