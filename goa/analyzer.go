package main

import (
	"fmt"
	"go/ast"
	"strings"
)

// Create API definition
// Extract resources, media types and controllers
func analyze(packages map[string]*ast.Package) (*apiDescription, errors) {
	description := newApiDescription()
	errs := new(errors)
	types := make(map[string]*ast.TypeSpec)

	// First pass, record all type definitions e.g. so we can find
	// payload struct definitions in second pass.
	for _, pkg := range packages {
		for _, object := range pkg.Scope.Objects {
			if object.Kind != ast.Typ {
				continue
			}
			spec := object.Decl.(*ast.TypeSpec)
			types[spec.Name.Name] = spec
		}
	}

	// Second pass, actually analyze relevant type definitions
	for _, pkg := range packages {
		for _, object := range pkg.Scope.Objects {
			if object.Kind != ast.Typ {
				continue
			}
			spec := object.Decl.(*ast.TypeSpec)
			docs := spec.Doc.List
			for _, d := range docs {
				if strings.HasPrefix(d.Text, "//@goa Resource") {
					if res, err := analyzeResource(spec, types); err != nil {
						errs.add(err)
					} else {
						errs.addIf(description.addResource(res))
					}
					break
				} else if strings.HasPrefix(d.Text, "//@goa MediaType") {
					if m, err := analyzeMediaType(spec, types); err != nil {
						errs.add(err)
					} else {
						errs.addIf(description.addMediaType(m))
					}
					break
				} else if strings.HasPrefix(d.Text, "//@goa Controller") {
					if c, err := analyzeController(spec); err != nil {
						errs.add(err)
					} else {
						errs.addIf(description.addController(c))
					}
					break
				} else if strings.HasPrefix(d.Text, "//@goa") {
					errs.add(fmt.Errorf("Unknown @goa directive '%s' for type declaration %s, first directive must be one of 'Resource', 'Controller' or 'MediaType'",
					d.Text, spec.Name.Name))
				}
			}
		}
	}

	return description, *errs
}

// TBD: Check that action parameters use JSON compatible types (numbers, bool or string)
func analyzeResource(spec *ast.TypeSpec, types map[string]*ast.TypeSpec) (*resourceDef, error) {
	for _, d := range spec.Doc.List {
		comment := d.Text
		if len(comment) < 8 {
			continue
		}
		if comment[:7] != "//@goa " {
			continue
		}
	}
	return &resourceDef{}, nil
}

func analyzeMediaType(spec *ast.TypeSpec, types map[string]*ast.TypeSpec) (*mediaTypeDef, error) {
	return &mediaTypeDef{}, nil
}

func analyzeController(spec *ast.TypeSpec) (*controllerDef, error) {
	return &controllerDef{}, nil
}
