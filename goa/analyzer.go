package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"mime"
	"regexp"
	"strconv"
	"strings"
)

// Top directive prefixes
const goaPrefix = "//@goa "
const resourcePrefix = "//@goa Resource"
const mediaTypePrefix = "//@goa MediaType:"
const controllerPrefix = "//@goa Controller:"

// Resource directive prefixes
const versionPrefix = "//@goa Version:"
const basePathPrefix = "//@goa BasePath:"

// Create API definition
// Extract resources, media types and controllers
func analyze(packages map[string]*ast.Package) (*apiDescription, errors) {
	// First pass, record all type definitions e.g. so we can find
	// payload struct definitions in second pass.
	types := make(map[string]*ast.TypeSpec)
	visitTypeSpecs(packages, func(spec *ast.TypeSpec) {
		types[spec.Name.Name] = spec
	})
	// Second pass, actually analyze relevant type definitions
	description := newApiDescription()
	errs := new(errors)
	visitTypeSpecs(packages, func(spec *ast.TypeSpec) {
		analyzeSpec(spec, description, errs)
	})
	return description, *errs
}

// Traverse packages and apply callback to all type specs
func visitTypeSpecs(packages map[string]*ast.Package, v func(*ast.TypeSpec)) {
	for _, pkg := range packages {
		for _, f := range pkg.Files {
			for _, d := range f.Decls {
				switch decl := d.(type) {
				case *ast.GenDecl:
					if decl.Tok != token.TYPE {
						continue
					}
					for _, s := range decl.Specs {
						v(s.(*ast.TypeSpec))
					}
				}
			}
		}
	}
}

// Check whether type spec has goa directives and if so create corresponding
// construct (resource, controller or media type).
func analyzeSpec(spec *ast.TypeSpec, description *apiDescription, errs *errors) {
	docs := spec.Doc
	if docs == nil {
		return
	}
	for _, d := range docs.List {
		if strings.HasPrefix(d.Text, resourcePrefix) {
			if res, err := analyzeResource(spec); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addResource(res))
			}
			break
		} else if strings.HasPrefix(d.Text, mediaTypePrefix) {
			if m, err := analyzeMediaType(spec, d.Text); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addMediaType(m))
			}
			break
		} else if strings.HasPrefix(d.Text, controllerPrefix) {
			if c, err := analyzeController(spec, d.Text); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addController(c))
			}
			break
		} else if strings.HasPrefix(d.Text, goaPrefix) {
			errs.add(fmt.Errorf("Unknown @goa directive '%s' for type declaration %s, first directive must start with '%s', '%s' or '%s'",
				d.Text, spec.Name.Name, resourcePrefix,
				mediaTypePrefix, controllerPrefix))
		}
	}
}

// TBD: Check that action parameters use JSON compatible types (numbers, bool or string)
func analyzeResource(spec *ast.TypeSpec) (*resourceDef, error) {
	resourceName := spec.Name.Name
	i, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("Resource %s must be an interface",
			resourceName)
	}
	version := ""
	mediaType := ""
	basePath := ""
	for _, d := range spec.Doc.List {
		text := d.Text
		if strings.HasPrefix(text, versionPrefix) &&
			len(text) > len(versionPrefix) {
			version = strings.Trim(text[len(versionPrefix):], " ")
		} else if strings.HasPrefix(text, mediaTypePrefix) &&
			len(text) > len(mediaTypePrefix) {
			mediaType = strings.Trim(text[len(mediaTypePrefix):], " ")
		} else if strings.HasPrefix(text, basePathPrefix) &&
			len(text) > len(basePathPrefix) {
			basePath = strings.Trim(text[len(basePathPrefix):], " ")
		} else if strings.HasPrefix(text, goaPrefix) {
			return nil, fmt.Errorf("Unknown goa directive for resource %s, resource directives must start with %s, %s or %s",
				resourceName, versionPrefix, mediaTypePrefix,
				basePathPrefix)
		}
	}
	if mediaType == "" {
		return nil, fmt.Errorf("Missing media type directive for resource %s, add a comment starting with %s", resourceName, mediaTypePrefix)
	}
	methods := i.Methods.List
	actionDefs := make(map[string]*actionDef, len(methods))
	for _, method := range methods {
		httpMethod := ""
		path := ""
		responses := make(map[int]*responseDef)
		actionName := method.Names[0].Name
		for _, d := range method.Comment.List {
			text := d.Text
			if strings.HasPrefix(text, goaPrefix) {
				if ms := methRegex.FindStringSubmatch(text); ms != nil {
					httpMethod = ms[1]
					path = ms[2]
				} else if ms = respRegex.FindStringSubmatch(text); ms != nil {
					code, err := strconv.Atoi(ms[1])
					if err != nil {
						return nil, fmt.Errorf("Invalid status code in %s for action %s of resource %s",
							ms[1], actionName, resourceName)
					}
					r, ok := responses[code]
					if !ok {
						r = &responseDef{code: code}
					}
					r.mediaType = ms[2]
				} else if ms = headerRegex.FindStringSubmatch(text); ms != nil {
					code, err := strconv.Atoi(ms[1])
					if err != nil {
						return nil, fmt.Errorf("Invalid status code in %s for action %s of resource %s",
							ms[1], actionName, resourceName)
					}
					r, ok := responses[code]
					if !ok {
						r = &responseDef{code: code}
					}
					r.headers[ms[2]] = ms[3]
				} else {
					return nil, fmt.Errorf("Unknown goa directive for action %s of resource %s, action directives must start with '//@goa <http method> <action path>', '//@goa <http status code>: [<response media type>]' or '//@goa <status code> <header name>: <header value or regex>'",
						actionName, resourceName)
				}
			}
		}
		if httpMethod == "" {
			return nil, fmt.Errorf("Missing path directive for action %s of resource %, add a comment starting with '//@goa <http method> \"<path>\"'",
				actionName, resourceName)
		}
		for _, r := range responses {
			mt, _, err := mime.ParseMediaType(r.mediaType)
			if err != nil {
				return nil, fmt.Errorf("Invalid media type identifier '%s' for action %s of resource %s (%s)",
					r.mediaType, actionName, resourceName,
					err.Error())
			}
			r.mediaType = mt
		}
		actionDefs[actionName] = &actionDef{
			name:      actionName,
			method:    httpMethod,
			path:      path,
			responses: responses,
			field:     method,
		}
	}
	return &resourceDef{
		name:       resourceName,
		apiVersion: version,
		basePath:   basePath,
		mediaType:  mediaType,
		actions:    actionDefs,
		spec:       spec,
	}, nil
}

func analyzeMediaType(spec *ast.TypeSpec, directive string) (*mediaTypeDef, error) {
	mediaTypeName := spec.Name.Name
	_, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("Media type %s must be a struct",
			mediaTypeName)
	}
	identifier := directive[len(mediaTypePrefix):]
	mt, _, err := mime.ParseMediaType(identifier)
	if err != nil {
		return nil, fmt.Errorf("Invalid media type identifier '%s' for media type %s (%s)",
			identifier, mediaTypeName, err.Error())
	}

	return &mediaTypeDef{mt, spec}, nil
}

func analyzeController(spec *ast.TypeSpec, directive string) (*controllerDef, error) {
	controllerName := spec.Name.Name
	_, ok := spec.Type.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("Controller %s must be a struct",
			controllerName)
	}
	resourceName := strings.Trim(directive[len(controllerPrefix):], " ")

	return &controllerDef{resourceName, spec}, nil
}

// Action directive regexps

var methRegex = regexp.MustCompile(
	"(GET|POST|PUT|DELETE|OPTIONS|HEAD|TRACE|CONNECT) \"(.*)\"")

var respRegex = regexp.MustCompile(
	"(100|101|200|201|202|203|204|205|206|300|301|302|303|304|305|307|" +
		"400|401|402|403|404|405|406|407|408|409|410|411|412|413|414|" +
		"415|416|417|418|500|501|502|503|504|505):\\s*(.*)")

var headerRegex = regexp.MustCompile(
	"(100|101|200|201|202|203|204|205|206|300|301|302|303|304|305|307|" +
		"400|401|402|403|404|405|406|407|408|409|410|411|412|413|414|" +
		"415|416|417|418|500|501|502|503|504|505) (.+):\\s*(.+)")
