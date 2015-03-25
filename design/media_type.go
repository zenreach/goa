package design

import (
	"fmt"
	"reflect"
	"strings"
)

// A media type describes the rendering of a resource using property and link
// definitions. An property corresponds to a single member of the media type,
// it has a name and a type as well as optional validation rules. A link has a
// name and a URL that points to a related resource.
// Finally media types also define views which describe which members and
// links to render when building the response body.
type MediaType struct {
	Object
	Identifier   string           // RFC 6838 Media type identifier
	Description  string           // Optional description
	Links        map[string]*Link // List of rendered links indexed by name (named hrefs to related resources)
	Views        map[string]*View // List of supported views indexed by name
	isCollection bool             // Whether media type is for a collection
}

// A link contains a URL to a related resource.
type Link struct {
	Name        string // Link name
	Description string // Optional description
	Member      string // Name of field used to render link if not Name
	View        string // View used to render link if not "link"
}

// A view defines which members and links to render when building a response.
// The view property names must match the names of the parent media type members.
// The members fields are inherited from the parent media type but may be overridden.
type View struct {
	Object
	Links []string
	Name  string
}

// NewMediaType creates new media type from its identifier, description and type.
// Initializes a default view that returns all the media type members.
func NewMediaType(id, desc string, o Object) *MediaType {
	mt := MediaType{Object: o, Identifier: id, Description: desc}
	defaultView := make(Object, len(o))
	for n, p := range o {
		defaultView[n] = p
	}
	mt.Views["default"] = &View{Name: "default", Object: defaultView}
	return &mt
}

// View adds a new view to the media type.
// It returns the view so it can be modified further.
// This method ignore passed-in property names that do not exist in media type.
func (m *MediaType) View(name string, members ...string) *View {
	o := make(Object, len(members))
	i := 0
	for n, p := range m.Object {
		found := false
		for _, m := range members {
			if m == n {
				found = true
				break
			}
		}
		if found {
			o[n] = p
			i += 1
		}
	}
	view := View{Name: name, Object: o}
	m.Views[name] = &view
	return &view
}

// As sets the list of member names rendered by view.
// If a member is a media type then the view used to render it defaults to the view with same name.
// The view used to renber media types members can be explicitely set using the syntax
// "<member name>:<view name>". For example:
//     m.View("expanded").As("id", "expensive_attribute:default")
func (v *View) As(members ...string) *View {
	o := Object{}
	for _, m := range members {
		o[m] = &Member{}
	}
	v.Object = o
	return v
}

// Links specifies the list of links rendered with this media type.
func (v *View) Link(links ...string) *View {
	v.Links = append(v.Links, links...)
	return v
}

// As sets the list of member names rendered by view

// Link adds a new link to the media type.
// It returns the link so it can be modified further.
func (m *MediaType) Link(name string) *Link {
	return &Link{Name: name, Member: name}
}

// Using sets the link Member field.
// It returns the link so it can be modified further.
func (l *Link) Using(member string) *Link {
	l.Member = member
	return l
}

// CollectionOf creates a collection media type from its element media type.
// A collection media type represents the content of responses that return a
// collection of resources such as "index" actions.
func CollectionOf(m *MediaType) *MediaType {
	col := MediaType{
		Object:       m.Object,
		Identifier:   m.Identifier,
		Description:  m.Description,
		Links:        m.Links,
		Views:        m.Views,
		isCollection: true,
	}
	return &col
}

// Render accepts either a struct or a map indexed by keys.
// If given a struct Render picks the struct fields whose names match the view property names.
// If the fields are tagged with json tags then Render uses the tag names to do the comparison with
// view property names.
// If given a map indexed by strings then Renders picks the keys with the same name as the view
// property names.
// If given an array then checks that media type is a collection then apply algorithm recursively
// on each element of the array.
// Once the resulting map has been built the values are validated using the view property
// validations.
func (m *MediaType) Render(value interface{}, viewName string) (interface{}, error) {
	if value == nil {
		return make(map[string]interface{}), nil
	}
	if _, ok := m.Views[viewName]; !ok {
		return nil, fmt.Errorf("View '%s' not found", viewName)
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		a := value.([]interface{})
		res := make([]interface{}, len(a))
		for i, e := range a {
			var err error
			if res[i], err = m.Render(e, viewName); err != nil {
				return nil, err
			}
		}
		return res, nil
	case reflect.Struct:
		return m.renderStruct(value, viewName)
	case reflect.Map:
		return m.renderMap(value.(map[string]interface{}), viewName)
	default:
		return nil, fmt.Errorf("Rendered value must be either a map or a struct. Given value was a %v.",
			reflect.TypeOf(value))
	}
}

// Render given struct
// Builds map with values corresponding to fields with media type property names then validates it
// View name must be valid
func (m *MediaType) renderStruct(value interface{}, viewName string) (map[string]interface{}, error) {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	numField := t.NumField()
	rendered := make(map[string]interface{})
	view := m.Views[viewName]
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			name = strings.Split(tag, ",")[0]
		}
		if _, ok := view.Object[name]; ok {
			rendered[name] = v.FieldByName(field.Name).Interface()
		}
	}
	if err := m.validate(rendered); err != nil {
		return nil, err
	}
	return rendered, nil
}

// Render given map
// Builds map with values corresponding to media type property names then validates it
// View name must be valid
func (m *MediaType) renderMap(value map[string]interface{}, viewName string) (map[string]interface{}, error) {
	rendered := make(map[string]interface{})
	view := m.Views[viewName]
	for n, v := range value {
		if _, ok := view.Object[n]; ok {
			rendered[n] = v
		}
	}
	if err := m.validate(rendered); err != nil {
		return nil, err
	}
	return rendered, nil
}

// First make sure that any property with default value has its value initialized in the map, then
// run property validation functions on value associated with property name.
func (m *MediaType) validate(rendered map[string]interface{}) error {
	for n, p := range m.Object {
		if _, ok := rendered[n]; !ok {
			if p.DefaultValue != nil {
				rendered[n] = p.DefaultValue
			}
		}
	}
	for n, v := range rendered {
		p := m.Object[n]
		for _, validate := range p.Validations {
			if err := validate(n, v); err != nil {
				return err
			}
		}
	}
	return nil
}
