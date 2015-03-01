package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"mime"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Top directive prefixes
	goaPrefix        = "@goa "
	resourcePrefix   = "@goa Resource"
	mediaTypePrefix  = "@goa MediaType:"
	controllerPrefix = "@goa Controller:"

	// Resource directive prefixes
	versionPrefix  = "@goa Version:"
	basePathPrefix = "@goa BasePath:"
	namePrefix     = "@goa Name:"

	// Action directive prefixes
	actionPrefix = "@goa Action:"
	viewsPrefix  = "@goa Views:"
)

var (
	// Action method directive
	methRegex = regexp.MustCompile(
		"(GET|POST|PUT|DELETE|OPTIONS|HEAD|TRACE|CONNECT) \"(.*)\"")

	// Action response directive
	respRegex = regexp.MustCompile(
		"(100|101|200|201|202|203|204|205|206|300|301|302|303|304|305|307|" +
			"400|401|402|403|404|405|406|407|408|409|410|411|412|413|414|" +
			"415|416|417|418|500|501|502|503|504|505):\\s*(.*)")

	// Action response header directive
	headerRegex = regexp.MustCompile(
		"(100|101|200|201|202|203|204|205|206|300|301|302|303|304|305|307|" +
			"400|401|402|403|404|405|406|407|408|409|410|411|412|413|414|" +
			"415|416|417|418|500|501|502|503|504|505) (.+):\\s*(.+)")
)

// Analyzer exposes methods to create resource, controller and media type
// definitions out of go AST packages.
type analyzer struct {
	packages map[string]*ast.Package
	verbose  bool
	docs     map[string]*doc.Package
}

// Factory method for analyzer
func newAnalyzer(packages map[string]*ast.Package, verbose bool) *analyzer {
	docs := make(map[string]*doc.Package, len(packages))
	for name, pkg := range packages {
		d := doc.New(pkg, "./", doc.AllDecls+doc.AllMethods)
		docs[name] = d
	}
	return &analyzer{packages, verbose, docs}
}

// Create API definition
// Extract resources, media types and controllers
func (a *analyzer) analyze() (*apiDescription, errors) {
	// First pass, record all type definitions e.g. so we can find
	// payload struct definitions in second pass.
	types := make(map[string]*doc.Type)
	a.visitTypes(func(spec *doc.Type) {
		types[spec.Name] = spec
	})
	// Second pass, actually analyze relevant type definitions
	description := newApiDescription()
	errs := new(errors)
	a.visitTypes(func(spec *doc.Type) {
		a.analyzeType(spec, description, errs)
	})
	return description, *errs
}

// Traverse packages and apply callback to all type specs
func (a *analyzer) visitTypes(v func(*doc.Type)) {
	for _, p := range a.docs {
		for _, t := range p.Types {
			v(t)
		}
	}
}

// Check whether type spec has goa directives and if so create corresponding
// construct (resource, controller or media type).
func (a *analyzer) analyzeType(spec *doc.Type, description *apiDescription, errs *errors) {
	docs := spec.Doc
	if docs == "" {
		return
	}
	for _, d := range strings.Split(docs, "\n") {
		if strings.HasPrefix(d, resourcePrefix) {
			if res, err := a.analyzeResource(spec); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addResource(res))
			}
			break
		} else if strings.HasPrefix(d, mediaTypePrefix) {
			if m, err := a.analyzeMediaType(spec, d); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addMediaType(m))
			}
			break
		} else if strings.HasPrefix(d, controllerPrefix) {
			if c, err := a.analyzeController(spec, d); err != nil {
				errs.add(err)
			} else {
				errs.addIf(description.addController(c))
			}
			break
		} else if strings.HasPrefix(d, goaPrefix) {
			errs.add(fmt.Errorf("Unknown @goa directive '%s' for type declaration %s, first directive must start with '%s', '%s' or '%s'",
				d, spec.Name, resourcePrefix,
				mediaTypePrefix, controllerPrefix))
		}
	}
}

// TBD: Check that action parameters use JSON compatible types (numbers, bool or string)
func (a *analyzer) analyzeResource(spec *doc.Type) (*ResourceDirective, error) {
	resourceName := spec.Name
	version := ""
	mediaType := ""
	basePath := ""
	for _, text := range strings.Split(spec.Doc, "\n") {
		text = strings.Trim(text, " ")
		if text == "@goa Resource" {
			continue
		}
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
			return nil, fmt.Errorf("Unknown goa directive '%s' for resource %s, resource directives must start with %s, %s or %s",
				text, resourceName, versionPrefix,
				mediaTypePrefix, basePathPrefix)
		}
	}
	if mediaType == "" {
		return nil, fmt.Errorf("Missing media type directive for resource %s, add a comment starting with %s", resourceName, mediaTypePrefix)
	}
	methods := spec.Methods
	ActionDefs := make(map[string]*ActionDirective, len(methods))
	for _, method := range methods {
		httpMethod := ""
		path := ""
		responses := make(map[int]*ResponseDirective)
		actionName := method.Name
		views := []string{}
		for _, text := range strings.Split(method.Doc, "\n") {
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
						r = &ResponseDirective{code: code}
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
						r = &ResponseDirective{code: code}
					}
					r.headers[ms[2]] = ms[3]
				} else if strings.HasPrefix(text, actionPrefix) &&
					len(text) > len(actionPrefix) {
					actionName = strings.Trim(text[len(actionPrefix):], " ")
				} else if strings.HasPrefix(text, viewsPrefix) &&
					len(text) > len(viewsPrefix) {
					views = strings.Split(strings.Trim(viewsPrefix, " "), ",")
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
		ActionDefs[actionName] = &ActionDirective{
			name:      actionName,
			method:    httpMethod,
			path:      path,
			responses: responses,
			views:     views,
			docs:      method,
		}
	}
	return &ResourceDirective{
		name:       resourceName,
		apiVersion: version,
		basePath:   basePath,
		mediaType:  mediaType,
		actions:    ActionDefs,
		docs:       spec,
	}, nil
}

func (a *analyzer) analyzeMediaType(spec *doc.Type, directive string) (*MediaTypeDirective, error) {
	mediaTypeName := spec.Name
	identifier := strings.Trim(directive[len(mediaTypePrefix):], " ")
	mt, _, err := mime.ParseMediaType(identifier)
	if err != nil {
		return nil, fmt.Errorf("Invalid media type identifier '%s' for media type %s (%s)",
			identifier, mediaTypeName, err.Error())
	}

	return &MediaTypeDirective{name: spec.Name, identifier: mt, docs: spec}, nil
}

func (a *analyzer) analyzeController(spec *doc.Type, directive string) (*ControllerDirective, error) {
	resourceName := strings.Trim(directive[len(controllerPrefix):], " ")
	return &ControllerDirective{spec.Name, resourceName, spec}, nil
}
