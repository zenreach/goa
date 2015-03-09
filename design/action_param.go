package design

// An action parameter (path element, query string or payload)
type ActionParam Property

// A validation takes a value and produces nil on success or an error otherwise
type Validation func(val interface{}) error

// A map of action parameters indexed by name
type ActionParams map[string]*ActionParam

// Null sets the action parameter type to Null
func (p *ActionParam) Null() *ActionParam {
	p.Type = Null
	return p
}

// Boolean sets the action parameter type to Boolean
func (p *ActionParam) Boolean() *ActionParam {
	p.Type = Boolean
	return p
}

// Integer sets the action parameter type to Integer
func (p *ActionParam) Integer() *ActionParam {
	p.Type = Integer
	return p
}

// Number sets the action parameter type to Number
func (p *ActionParam) Number() *ActionParam {
	p.Type = Number
	return p
}

// String sets the action parameter type to String
func (p *ActionParam) String() *ActionParam {
	p.Type = String
	return p
}

// Array sets the action parameter type to Array
func (p *ActionParam) Array(elemType DataType) *ActionParam {
	p.Type = &Array{ElemType: elemType}
	return p
}

// Object sets the action parameter type to Object
func (p *ActionParam) Object(blueprint interface{}, properties ...*Property) *ActionParam {
	object := make(Object, len(properties))
	for _, p := range properties {
		object[p.Name] = p
	}
	p.Type = &object
	return p
}
