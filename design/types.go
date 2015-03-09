package design

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// A Kind represents the specific kind of type that a DataType represents.
type Kind uint

const (
	NullType Kind = iota
	BooleanType
	IntegerType
	NumberType
	StringType
	ArrayType
	ObjectType
)

// DataType interface represents both JSON schema types and media types.
// All data types have a kind (Integer, Number etc. for JSON schema types
// and Object for media types) and a "Load" method.
// The "Load" method checks that the value of the given argument is compatible
// with the type and returns the coerced value if that's case, an error otherwise.
// Data types are used to define the type of media type propertys and of action
// parameters.
type DataType interface {
	Kind() Kind                               // integer, number, string, ...
	Name() string                             // Human readable name
	Load(interface{}) (interface{}, error)    // Validate and load
	CanLoad(t reflect.Type, ctx string) error // nil if values of given type can be loaded into fields described by attribute, descriptive error otherwise
}

// Type for null, boolean, integer, number and string
type Primitive Kind

var (
	// Type for the JSON null value
	Null = Primitive(NullType)
	// Type for a JSON boolean
	Boolean = Primitive(BooleanType)
	// Type for a JSON number without a fraction or exponent part
	Integer = Primitive(IntegerType)
	// Type for any JSON number, including integers
	Number = Primitive(NumberType)
	// Type for a JSON string
	String = Primitive(StringType)
)

// Type kind
func (b Primitive) Kind() Kind {
	return Kind(b)
}

// Human readable name of basic type
func (b Primitive) Name() string {
	switch Kind(b) {
	case NullType:
		return "null"
	case BooleanType:
		return "boolean"
	case IntegerType:
		return "integer"
	case NumberType:
		return "number"
	case StringType:
		return "string"
	default:
		panic(fmt.Sprintf("goa bug: unknown basic type %#v", b))
	}
}

// Attempt to load value into basic type
// How a value is coerced depends on its type and the basic type kind:
// Only strings may be loaded in values of type String.
// Any integer value or string representing an integer may be loaded in values of type Integer.
// Any integer or float value or string representing integers or floats may be loaded in values of
// type Number.
// true, false, 1, 0, "false", "FALSE", "0", "f", "F", "true", "TRUE", "1", "t", "T" may be loaded
// in values of type Boolean.
// Returns nil and an error if coercion fails.
func (b Primitive) Load(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	var extra string
	switch Kind(b) {
	case BooleanType:
		switch v := value.(type) {
		case bool:
			return value, nil
		case string:
			if res, err := strconv.ParseBool(v); err == nil {
				return res, nil
			} else {
				extra = err.Error()
			}
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			if value == 0 {
				return false, nil
			} else if value == 1 {
				return true, nil
			} else {
				extra = "integer value must be 0 or 1"
			}
		}
	case IntegerType:
		switch v := value.(type) {
		case int:
			return v, nil
		case uint:
			return int(v), nil
		case int8:
			return int(v), nil
		case uint8:
			return int(v), nil
		case int16:
			return int(v), nil
		case uint16:
			return int(v), nil
		case int32:
			return int(v), nil
		case uint32:
			return int(v), nil
		case int64:
			return int(v), nil
		case uint64:
			return int(v), nil
		case string:
			if res, err := strconv.ParseInt(v, 10, 0); err == nil {
				return int(res), nil
			} else {
				extra = err.Error()
			}
		}
	case NumberType:
		switch v := value.(type) {
		case int:
			return float64(v), nil
		case uint:
			return float64(v), nil
		case int8:
			return float64(v), nil
		case uint8:
			return float64(v), nil
		case int16:
			return float64(v), nil
		case uint16:
			return float64(v), nil
		case int32:
			return float64(v), nil
		case uint32:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case uint64:
			return float64(v), nil
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case string:
			if res, err := strconv.ParseFloat(v, 64); err == nil {
				return res, nil
			} else {
				extra = err.Error()
			}
		}
	case StringType:
		if _, ok := value.(string); ok {
			return value, nil
		}
	}
	return nil, &IncompatibleValue{value: value, to: b.Name(), extra: extra}
}

// CanLoad checks whether values of the given go type can be loaded into values of this basic goa type.
// Returns nil if check is successful, error otherwise.
func (b Primitive) CanLoad(t reflect.Type, context string) error {
	switch Kind(b) {
	case BooleanType:
		switch t.Kind() {
		case reflect.Bool, reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16,
			reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64, reflect.String:
			return nil
		}
	case IntegerType:
		switch t.Kind() {
		case reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32,
			reflect.Uint32, reflect.Int64, reflect.Uint64:
			return nil
		}
	case NumberType:
		if t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64 {
			return nil
		}
	case StringType:
		if t.Kind() == reflect.String {
			return nil
		}
	}
	return &IncompatibleType{context: context, to: t}
}

// An array of values of type ElemType
type Array struct {
	ElemType DataType
}

// Type kind
func (a *Array) Kind() Kind {
	return ArrayType
}

// Load coerces the given value into a []interface{} where the array values have all been coerced recursively.
// `value` must either be a slice, an array or a string containing a JSON representation of an array.
// Load also applies any validation rule defined in the array element properties.
// Returns nil and an error if coercion or validation fails.
func (a *Array) Load(value interface{}) (interface{}, error) {
	var arr []interface{}
	k := reflect.TypeOf(value).Kind()
	if k == reflect.String {
		if err := json.Unmarshal([]byte(value.(string)), &arr); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Array",
				extra: fmt.Sprintf("failed to decode JSON: %v", err.Error())}
		}
	} else if k == reflect.Slice || k == reflect.Array {
		v := reflect.ValueOf(value)
		for i := 0; i < v.Len(); i++ {
			arr = append(arr, v.Index(i).Interface())
		}
	} else {
		return nil, &IncompatibleValue{value: value, to: "Array",
			extra: "value must be an array or a slice"}
	}
	var res []interface{}
	varr := reflect.ValueOf(arr)
	for i := 0; i < varr.Len(); i++ {
		ev, err := a.ElemType.Load(varr.Index(i).Interface())
		if err != nil {
			return nil, &IncompatibleValue{value: value, to: "Array",
				extra: fmt.Sprintf("cannot load value at index %v: %v", i, err.Error())}
		}
		res = append(res, ev)
	}
	return interface{}(res), nil
}

// CanLoad checks whether values of the given go type can be loaded into values of this array.
// Returns nil if check is successful, error otherwise.
func (a *Array) CanLoad(t reflect.Type, context string) error {
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return &IncompatibleType{context: context, to: t, extra: "value must be an array or a slice"}
	}
	return a.ElemType.CanLoad(t.Elem(), fmt.Sprintf("%v items", context))
}

// JSON schema type name
func (a *Array) Name() string {
	return "array"
}

// A JSON object
type Object map[string]*Property

// NewObjectType creates a new object type from the given properties.
func NewObject(properties ...*Property) Object {
	o := make(Object, len(properties))
	o.Init(properties...)
	return o
}

// Init initializes the object properties
func (o Object) Init(properties ...*Property) {
	for _, p := range properties {
		o[p.Name] = p
	}
}

// Type kind
func (o Object) Kind() Kind {
	return ObjectType
}

// Load coerces the given value into a map[string]interface{} where the map values have all been coerced recursively.
// `value` must either be a map with string keys or to a string containing a JSON representation of a map.
// Load also applies any validation rule defined in the object properties.
// Returns `nil` and an error if coercion or validation fails.
func (o Object) Load(value interface{}) (interface{}, error) {
	// First load from JSON if needed
	var m map[string]interface{}
	switch value.(type) {
	case string:
		if err := json.Unmarshal([]byte(value.(string)), &m); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Object", extra: "string is not a JSON object"}
		}
	case map[string]interface{}:
		m = value.(map[string]interface{})
	default:
		return nil, &IncompatibleValue{value: value, to: "Object"}
	}
	// Now go through each type member and load and validate value from map
	coerced := make(map[string]interface{})
	var errors []error
	for n, prop := range o {
		val, ok := m[n]
		if !ok {
			if prop.DefaultValue != nil {
				val = prop.DefaultValue
			}
		} else {
			var err error
			val, err = prop.Type.Load(val)
			if err != nil {
				errors = append(errors, &IncompatibleValue{value,
					"Object",
					fmt.Sprintf("could not load property %s: %s", n,
						err.Error())})
				continue
			}
		}
		for _, validate := range prop.Validations {
			if err := validate(val); err != nil {
				errors = append(errors, err)
				continue
			}
		}
		coerced[n] = val
	}
	if len(errors) > 0 {
		// TBD create MultiError type
		return nil, errors[0]
	}
	return coerced, nil
}

// CanLoad checks whether values of the given go type can be loaded into values of object.
// Returns nil if check is successful, error otherwise.
func (o Object) CanLoad(t reflect.Type, context string) error {
	if t.Kind() != reflect.Struct {
		return &IncompatibleType{context: context, to: t, extra: "value must be a struct"}
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		name := f.Tag.Get("property")
		if len(name) == 0 {
			name = f.Name
		}
		prop, ok := o[name]
		newContext := fmt.Sprintf("%s.%v", context, f.Name)
		if !ok {
			return &IncompatibleType{context: newContext, to: t, extra: "No property with name " + f.Name}
		} else {
			if err := prop.Type.CanLoad(f.Type, newContext); err != nil {
				return err
			}
		}
	}
	return nil
}

// JSON schema type name
func (a Object) Name() string {
	return "object"
}

// An object property with optional description, default value and validations
type Property struct {
	Name         string       // Property name
	Type         DataType     // Property type
	Description  string       // Optional description
	Validations  []Validation // Optional validation functions
	DefaultValue interface{}  // Optional property default value
}

// Create new property from name and type
func Prop(n string, t DataType, desc string) *Property {
	return &Property{Name: n, Description: desc, Type: t}
}

// Create array property
func ArrayProp(n, desc string, elemType DataType) *Property {
	return &Property{Name: n, Description: desc, Type: &Array{ElemType: elemType}}
}

// Create object property
func ObjectProp(n, desc string, properties ...*Property) *Property {
	object := make(Object, len(properties))
	for _, p := range properties {
		object[p.Name] = p
	}
	return &Property{Name: n, Description: desc, Type: object}
}

// Error raised when "Load" cannot coerce a value to the data type
type IncompatibleValue struct {
	value interface{} // Value being loaded
	to    string      // Name of type being coerced to
	extra string      // Extra error information if any
}

// Error returns the error message
func (e *IncompatibleValue) Error() string {
	extra := ""
	if len(e.extra) > 0 {
		extra = ": " + e.extra
	}
	return fmt.Sprintf("Cannot load value %v into a %v%s", e.value, e.to, extra)
}

// Error raised when a values of given go type cannot be assigned to property's type (by `CanLoad()`)
type IncompatibleType struct {
	context string
	to      reflect.Type
	extra   string // Extra error information if any
}

// Error returns the error message
func (e *IncompatibleType) Error() string {
	extra := ""
	if len(e.extra) > 0 {
		extra = ": " + e.extra
	}
	prefix := ""
	if len(e.context) > 0 {
		prefix = e.context + " "
	}
	return fmt.Sprintf(prefix+"cannot be assigned values of type %v%s",
		e.to, extra)
}
