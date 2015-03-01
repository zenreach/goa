package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// Generator struct exposes methods to generate API code and documentation.
type generator struct {
	api *apiDescription
}

// Generator factory
func newGenerator(api *apiDescription) *generator {
	return &generator{api}
}

// Generator entry point: generate code for API
func (g *generator) generateApi(w io.Writer) errors {
	g.generateHeader(w)
	errs := errors{}
	names := make([]string, len(g.api.resources))
	idx := 0
	for name, _ := range g.api.resources {
		names[idx] = name
		idx += 1
	}
	sort.Strings(names)
	identifiers := make(map[string]bool, len(g.api.mediaTypes))
	for i, _ := range g.api.mediaTypes {
		identifiers[i] = true
	}
	for _, name := range names {
		resource, _ := g.api.resources[name]
		g.generateResource(resource, w)
		c, _ := g.api.resourceCompiler(resource)
		errs.addIf(g.generateController(c, w))
		delete(identifiers, resource.mediaType)
		errs.addIf(g.generateMediaType(resource.mediaType, w))
	}
	for i, _ := range identifiers {
		errs.addIf(g.generateMediaType(i, w))
	}
	return errs
}

func (g *generator) generateHeader(o io.Writer) {
	// pwd, _ := os.Getwd()
	// title := filepath.Base(pwd)
	// w := newWriter()
}

func (g *generator) generateMediaType(id string, o io.Writer) error {
	return nil
}

func (g *generator) generateResource(r *ResourceDirective, o io.Writer) error {
	schema, err := g.generateJsonSchema(g.api.mediaTypes[r.mediaType])
	if err != nil {
		return err
	}
	source, _ := json.MarshalIndent(schema, "", "    ")
	w := newWriter()
	w.w("//== %s ==\n\n", r.name)
	w.w("type goa_%s ResourceDefinition\n\n", r.name)
	w.w("var goa_%sSchema = %s\n\n", g.api.mediaTypes[r.mediaType].name, source)
	w.w("func goa_Mount%sHandlers(app goa.Application) {\n", r.name)

	w.flush(o)
	return nil
}

func (g *generator) generateController(controller *ControllerDirective, o io.Writer) error {
	return nil
}

// Generate JSON schema from arbitrary data structure.
// Struct field tags may be used to specify validation rules.
func (g *generator) generateJsonSchema(m *MediaTypeDirective) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
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
