package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"reflect" // Only to parse struct tags
	"regexp"
	"strconv"
	"strings"
)

var (
	// Valid "format" tag values
	validFormats = map[string]bool{
		"int":              true,
		"email":            true,
		"time.ANSIC":       true,
		"time.UnixDate":    true,
		"time.RubyDate":    true,
		"time.RFC822":      true,
		"time.RFC822Z":     true,
		"time.RFC850":      true,
		"time.RFC1123":     true,
		"time.RFC1123Z":    true,
		"time.RFC3339":     true,
		"time.RFC3339Nano": true,
		"time.Kitchen":     true,
		"time.Stamp":       true,
		"time.StampMilli":  true,
		"time.StampMicro":  true,
		"time.StampNano":   true,
	}
)

// Mediatype definition: defines identifier
type MediaTypeDirective struct {
	name         string                       // go struct name
	identifier   string                       // Media type identifier
	docs         *doc.Type                    // Documentation
	views        map[string][]string          // Media type views
	viewMappings map[string]map[string]string // Media type view mappings
}

// Produce JSON schema from media type node
func (m *MediaTypeDirective) build() error {
	m.views = make(map[string][]string)
	m.viewMappings = make(map[string]map[string]string)
	specs := m.docs.Decl.Specs
	if len(specs) > 1 {
		return fmt.Errorf("Invalid media type definition %s: more than one declaration.",
			m.name)
	}
	typeSpec, ok := specs[0].(*ast.TypeSpec)
	if !ok {
		return fmt.Errorf("Invalid media type definition %s: must be a type declaration.",
			m.name)
	}
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("Invalid media type definition %s: must be a struct declaration.",
			m.name)
	}
	var schema map[string]interface{}
	for _, field := range structType.Fields.List {
		name := field.Names[0].Name
		t := field.Tag.Value
		if len(t) > 0 {
			stag := reflect.StructTag(t)
			values := stag.Get("goa")
			if values != "" {
				vals := strings.Split(values, ",")
				for _, v := range vals {
					elems := strings.SplitN(v, ":", 2)
					gen, err := m.generator(elems[0])
					if err != nil {
						return fmt.Errorf("Invalid tag '%s': %s",
							err.Error())
					}
					tag, err := gen(elems[0], elems[1])
					if err != nil {
						return err
					}
					switch tag[0] {
					case "views":
						views := tag[1].([]string)
						for _, v := range views {
							if _, ok := m.views[v]; !ok {
								m.views[v] = []string{name}
							} else {
								m.views[v] = append(m.views[v], name)
							}
						}
					case "viewMappings":
						mappings := tag[1].(map[string]string)
						if _, ok := m.viewMappings[name]; !ok {
							m.viewMappings[name] = mappings
						} else {
							for n, p := range mappings {
								if _, ok := m.viewMappings[name][n]; ok {
									return fmt.Errorf("Duplicate view mapping definition for field %s view %s of media type %s",
										name, n, m.name)
								}
								m.viewMappings[name][n] = p
							}
						}
					}
					// HANDLE OTHER TAGS (build JSON SCHEMA)

				}
			}
		}
		for _, name := range field.Names {

		}

	}
	return nil
}

// Resource directives: version and default media type
// Interface that defines resource actions
type ResourceDirective struct {
	name       string                      // go interface name
	apiVersion string                      // API version - can be empty
	mediaType  string                      // Media type identifier
	basePath   string                      // Base path for all actions - can be empty
	actions    map[string]*ActionDirective // Resource action definitions
	docs       *doc.Type                   // Documentation
}

// Resource action directives: route and responses
type ActionDirective struct {
	name      string                     // Action name (method name)
	method    string                     // Action HTTP method ("GET", "POST", etc.)
	path      string                     // Action path relative to resource base path
	responses map[int]*ResponseDirective // Response definitions
	views     []string                   // Available views
	docs      *doc.Func                  // Documentation
}

// Response directives: body and headers
type ResponseDirective struct {
	code      int               // HTTP status code
	mediaType string            // Media type identifier
	headers   map[string]string // HTTP headers
}

// Controller directive: specifies resource being implemented
type ControllerDirective struct {
	name     string    // go struct name
	resource string    // Resource interface name
	docs     *doc.Type // Documentation
}

// Tag name and value
type tag struct {
	name  string
	value interface{}
}

// A tag generator accepts a tag name and value and returns an error if
// validation fails, nil otherwise.
type tagGenerator func(name, value string) (tag, error)

// Retrieve generator for given tag name, nil if no additional validation is
// needed.
// Return error if tag name is invalid.
func (m *MediaTypeDirective) generator(name string) (tagGenerator, error) {
	switch name {
	case "default":
		return m.stringTagGenerator, nil
	case "enum":
		return m.listTagGenerator, nil
	case "format":
		return m.stringTagGenerator, nil
	case "maxLength":
		return m.intTagGenerator, nil
	case "minLength":
		return m.intTagGenerator, nil
	case "maxValue":
		return m.intTagGenerator, nil
	case "minValue":
		return m.intTagGenerator, nil
	case "pattern":
		return m.stringTagGenerator, nil
	case "required":
		return m.unaryTagGenerator, nil
	case "views":
		return m.listTagGenerator, nil
	case "viewMappings":
		return m.mappingsTagGenerator, nil
	default:
		return nil, fmt.Errorf("Unknown tag '%s', valid tags are Enum, Format, MaxLength, MinLength, MaxValue, MinValue, Patter, Required, Views and ViewMappings", name)
	}
}

// Generate list tag value
func (m *MediaTypeDirective) listTagGenerator(name, value string) (tag, error) {
	if len(value) == 0 {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value cannot be empty.",
			name, value, m.name)
	}
	val := strings.Split(value, " ")
	return tag{name, val}, nil
}

// Generate unary tag
func (m *MediaTypeDirective) unaryTagGenerator(name, value string) (tag, error) {
	if len(value) > 0 {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value must be empty.",
			name, value, m.name)
	}
	return tag{name, nil}, nil
}

// Generate regex tag value
func (m *MediaTypeDirective) stringTagGenerator(name, value string) (tag, error) {
	if len(value) == 0 {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value cannot be empty.",
			name, value, m.name)
	}
	switch name {
	case "pattern":
		if _, err := regexp.Compile(value); err != nil {
			return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value must be a valid regular expression (%s)",
				name, value, m.name, err.Error())
		}
	case "format":
		if _, ok := validFormats[value]; !ok {
			keys := make([]string, len(validFormats))
			idx := 0
			for k, _ := range validFormats {
				keys[idx] = k
				idx += 1
			}
			return tag{}, fmt.Errorf("Invalid Format tag value '%v' on media type %s: value must be one of %s.",
				value, m.name, strings.Join(keys, ", "))
		}
	}
	return tag{name, value}, nil
}

// Generate mappings tag value
func (m *MediaTypeDirective) mappingsTagGenerator(name, value string) (tag, error) {
	if len(value) == 0 {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value cannot be empty.",
			name, value, m.name)
	}
	if len(strings.Split(value, "="))%2 == 1 {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s:  value must consist of a list of strings each containing the character '='.",
			name, value, m.name)
	}
	mappings := strings.Split(value, " ")
	val := make(map[string]string)
	for _, p := range mappings {
		elems := strings.Split(p, "=")
		if len(elems) != 2 {
			return tag{}, fmt.Errorf("Invalid mapping '%s' on media type %s, mapping syntax is <parent>=<child> e.g. 'default=tiny'.",
				p, m.name)
		}
		if _, ok := val[elems[0]]; ok {
			return tag{}, fmt.Errorf("Invalid mapping '%s' on media type %s, key '%s' appears twice.",
				value, m.name, elems[0])
		}
		val[elems[0]] = elems[1]
	}
	return tag{name, val}, nil
}

// Generate integer tag value
func (m *MediaTypeDirective) intTagGenerator(name, value string) (tag, error) {
	val, err := strconv.Atoi(value)
	if err != nil {
		return tag{}, fmt.Errorf("Invalid %s tag value '%v' on media type %s: value must be an integer.",
			name, value, m.name)
	}
	return tag{name: val}, nil
}
