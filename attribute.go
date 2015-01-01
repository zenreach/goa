package goa

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Attributes are used to describe data structures: an attribute defines a field name and type. It allows providing
// a description as well as type-specific validation rules (regular expression for strings, min and max values for
// integers etc.).
//
// The type of a field defined by an attribute may be one of 5 basic types (string, integer, float, boolean or time) or
// may be a composite type: another data structure also defined with attributes. A type may also be a collection (in
// which case it defines the type of the elements) or a hash (in which case it defines the type of the values, keys are
// always strings).
//
// All types (basic, composite, collection and hash) implement the `Type` interface. This interface exposes the `Load()`
// function which accepts the JSON representation of a value as well as other compatible representations (e.g. `int8`,
// `int16`, etc. for integers, "1", "true" for booleans etc.). This makes it possible to call it recursively on
// embedded data structures and coerce all fields to the type defined by the attributes (or fail if the value cannot be
// coerced). The Type interface also exposes a `CanLoad()` function which takes a go type (`reflect.Type`) and returns
// an error if values of that type cannot be coerced into the goa type or nil otherwise. The idea here is that `CanLoad`
// may be used to validate that a go structure matches an attribute definition. `Load()` can then be called multiple
// times to load values into instances of that structure.
//
// Validation rules apply whenever a value is loaded via the `Load()` method. They specify whether a field is required,
// regular expressions (for string attributes), minimum and maximum length (strings and collections) or minimum and
// maximum values (for integer, float and time attributes).
//
// Finally, attributes may also define a default value and/or a list of allowed values for a field.
//
// Here is an example of an attribute definition using a composite type:
//
//    article := Attribute{
//        Description: "An article",
//        Type: Composite{
//            "title": Attribute{
//                Type:        String,
//                Description: "Article title",
//                MinLength:   20,
//                MaxLength:   200,
//                Required:    true,
//            },
//            "author": Attribute{
//                Type: Composite{
//                    "firstName": Attribute{
//                        Type:        String,
//                        Description: "Author first name",
//                    },
//                    "lastName": Attribute{
//                        Type:        String,
//                        Description: "Author last name",
//                    },
//                },
//                Required: true,
//            },
//            "published": Attribute{
//                Type:        Time,
//                Description: "Article publication date",
//                Required:    true,
//            },
//        },
//    }
//
// The example above could represent values such as:
//
//   articleData := map[string]interface{}{
//       "title": "goa, a novel go web application framework",
//       "author": map[string]interface{}{
//           "firstName": "Leeroy",
//           "lastName":  "Jenkins",
//       },
//       "published": time.Now(),
//   }
//
type Attribute struct {
	Type         Type        // Attribute type
	Description  string      // Attribute description
	DefaultValue interface{} // Attribute default value (if any), underlying (go) type is dictated by `Type`

	// - Validation rules -
	Required      bool        // Whether the attribute is required when loading a value of this type
	Regexp        string      // Regular expression used to validate string values
	MinValue      interface{} // Minimum value used to validate integer, float and time values
	MaxValue      interface{} // Maximum value used to validate integer, float and time values
	MinLength     int         // Minimum value length used to validate strings and collections
	MaxLength     int         // Maximum value length used to validate strings and collections
	AllowedValues interface{} // White list of possible values, underlying type is an array
}

// Validate checks that the given attribute struct is properly initialized
func (a *Attribute) Validate() error {
	if a.Type == nil {
		return errors.New("field 'Type' is missing")
	}
	return nil
}

// Attribute kind
type Kind int

// List of supported kinds
const (
	//	Kind                   Go type produced by Load()
	TString     Kind = iota // string
	TInteger                // int
	TFloat                  // float64
	TBoolean                // bool
	TTime                   // time.Time
	TComposite              // map[string]interface{}
	TCollection             // []interface{}
	THash                   // map[string]interface{}
	_TLast                  // (none) _TLast is a special marker that contains the next value for the Kind enum
)

// Interface implemented by all types (basic, composite, collection and hash)
type Type interface {
	GetKind() Kind                                // Type kind, one of constants defined above
	Load(value interface{}) (interface{}, error)  // Load value, return error if `CanLoad()` would or a validation fails
	CanLoad(t reflect.Type, context string) error // nil if values of given type can be loaded into fields described by attribute, descriptive error otherwise
}

//** Types **/

// Basic types
type basic Kind

// String basic type
var String = basic(TString)

// Integer basic type
var Integer = basic(TInteger)

// Float basic type
var Float = basic(TFloat)

// Boolean basic type
var Boolean = basic(TBoolean)

// Time basic type
var Time = basic(TTime)

// Attributes map
type Attributes map[string]Attribute

// Composite type i.e. attributes map
type Composite Attributes

// Hash type
type Hash struct{ ElemType Type }

// Collection type
type Collection struct{ ElemType Type }

// CollectionOf creates a collection (array) type.
// Takes type of elements as argument.
func CollectionOf(t Type) Type {
	return &Collection{t}
}

// HashOf creates a hash type.
// Takes type of keys and values as argument, hash keys are always strings.
func HashOf(t Type) Type {
	return &Hash{t}
}

//** Load Error **/

// Error raised when a value cannot be coerced to attribute's type (by `Load()`)
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
	return fmt.Sprintf("Cannot load %v into a %v%s (got value %+v)", reflect.TypeOf(e.value), e.to, extra, e.value)
}

// Error raised when a values of given go type cannot be assigned to attribute's type (by `CanLoad()`)
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
	return fmt.Sprintf(prefix + "cannot be assigned values of type %v%s", 
		e.to, extra)
}

//** basic types implementation **/

// CanLoad checks whether values of the given go type can be loaded into values of this basic goa type.
// Returns nil if check is successful, error otherwise.
func (b basic) CanLoad(t reflect.Type, context string) error {
	switch Kind(b) {
	case TString:
		if t.Kind() == reflect.String {
			return nil
		}
	case TInteger:
		switch t.Kind() {
		case reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16, reflect.Int32,
			reflect.Uint32, reflect.Int64, reflect.Uint64:
			return nil
		}
	case TFloat:
		if t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64 {
			return nil
		}
	case TBoolean:
		switch t.Kind() {
		case reflect.Bool, reflect.Int, reflect.Uint, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16,
			reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64, reflect.String:
			return nil
		}
	case TTime:
		switch t.Kind() {
		case reflect.Struct, reflect.String: // time.Time Kind's is Struct
			return nil
		}
	}
	return &IncompatibleType{context: context, to: t}
}

// Load coerces value into this basic type.
// How a value is coerced depends on its type and the basic type kind:
// - Only strings may be loaded in attributes of type String.
// - Any go integer type value or string representing an integer may be loaded in attributes of type Integer.
// - Any go integer or float type value or string representing integers or floats may be loaded in attributes of
//   type Float.
// - true, false, 1, 0, "false", "FALSE", "0", "f", "F", "true", "TRUE", "1", "t", "T" may be loaded in attributes
//   of type Boolean.
// - Any go integer type value (unix time), time.Time value or string representing a time value may be loaded
//   in attributes of type Time.
// Time values given as string are parsed using the following standard (in attempt order): RFC3339, ANSIC, UnixDate,
// RubyDate, RFC822, RFC822Z, RFC850, RFC1123, RFC1123Z and RFC3339Nano, see time.Time for details.
// Returns nil and an error if coercion fails.
func (b basic) Load(value interface{}) (interface{}, error) {
	extra := ""
	switch Kind(b) {
	case TString:
		switch value.(type) {
		case string:
			return value.(string), nil
		}
	case TInteger:
		switch value.(type) {
		case int:
			return value.(int), nil
		case uint:
			return int(value.(uint)), nil
		case int8:
			return int(value.(int8)), nil
		case uint8:
			return int(value.(uint8)), nil
		case int16:
			return int(value.(int16)), nil
		case uint16:
			return int(value.(uint16)), nil
		case int32:
			return int(value.(int32)), nil
		case uint32:
			return int(value.(uint32)), nil
		case int64:
			return int(value.(int64)), nil
		case uint64:
			return int(value.(uint64)), nil
		case string:
			if res, err := strconv.ParseInt(value.(string), 10, 0); err == nil {
				return int(res), nil
			}
		}
	case TFloat:
		switch value.(type) {
		case int:
			return float64(value.(int)), nil
		case uint:
			return float64(value.(uint)), nil
		case int8:
			return float64(value.(int8)), nil
		case uint8:
			return float64(value.(uint8)), nil
		case int16:
			return float64(value.(int16)), nil
		case uint16:
			return float64(value.(uint16)), nil
		case int32:
			return float64(value.(int32)), nil
		case uint32:
			return float64(value.(uint32)), nil
		case int64:
			return float64(value.(int64)), nil
		case uint64:
			return float64(value.(uint64)), nil
		case float32:
			return float64(value.(float32)), nil
		case float64:
			return value.(float64), nil
		case string:
			if res, err := strconv.ParseFloat(value.(string), 64); err == nil {
				return res, nil
			}
		}
	case TBoolean:
		switch value.(type) {
		case bool:
			return value.(bool), nil
		case string:
			sm := *boolStringMap()
			if val, ok := sm[value.(string)]; ok {
				return val, nil
			} else {
				keys := make([]string, len(sm))
				i := 0
				for k, _ := range sm {
					keys[i] = k
					i++
				}
				extra = fmt.Sprintf("string value must be one of %v", strings.Join(keys, ", "))
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
	case TTime:
		switch value.(type) {
		case time.Time:
			return value.(time.Time), nil
		case string:
			layouts := []string{time.RFC3339, time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z, time.RFC850,
				time.RFC1123, time.RFC1123Z, time.RFC3339Nano}
			for _, layout := range layouts {
				if res, err := time.Parse(layout, value.(string)); err == nil {
					return res, nil
				}
			}
			extra = "time value must be formatted using ANSI C, Unix date, Ruby date, RFC822, RFC822Z, RFC850, RFC1123, RFC1123Z or RFC3389 standard"
		case int:
			return time.Unix(int64(value.(int)), 0), nil
		case uint:
			return time.Unix(int64(value.(uint)), 0), nil
		case int8:
			return time.Unix(int64(value.(int8)), 0), nil
		case uint8:
			return time.Unix(int64(value.(uint8)), 0), nil
		case int16:
			return time.Unix(int64(value.(int16)), 0), nil
		case uint16:
			return time.Unix(int64(value.(uint16)), 0), nil
		case int32:
			return time.Unix(int64(value.(int32)), 0), nil
		case uint32:
			return time.Unix(int64(value.(uint32)), 0), nil
		case int64:
			return time.Unix(int64(value.(int64)), 0), nil
		case uint64:
			return time.Unix(int64(value.(uint64)), 0), nil
		}
	}

	return nil, &IncompatibleValue{value: value, to: b.String(), extra: extra}
}

// GetKind returns the kind of this basic type (string, integer etc.)
func (b basic) GetKind() Kind {
	return Kind(b)
}

// String returns the name of this basic type ("String", "Integer", etc.)
func (b basic) String() string {
	switch Kind(b) {
	case TString:
		return "String"
	case TInteger:
		return "Integer"
	case TBoolean:
		return "Boolean"
	case TFloat:
		return "Float"
	case TTime:
		return "Time"
	default:
		return "??"
	}
}

// boolStringMap returns a mapping of strings to boolean values used when coercing strings into a boolean basic type
func boolStringMap() *map[string]bool {
	return &map[string]bool{"false": false, "FALSE": false, "0": false, "f": false, "F": false,
		"true": true, "TRUE": true, "1": true, "t": true, "T": true}
}

// Composite

// CanLoad checks whether values of the given go type can be loaded into values of this composite type.
// Returns nil if check is successful, error otherwise.
func (c Composite) CanLoad(t reflect.Type, context string) error {
	if t.Kind() != reflect.Struct {
		return &IncompatibleType{context: context, to: t, extra: "value must be a struct"}
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.FieldByIndex([]int{i})
		attName := f.Tag.Get("attribute")
		if len(attName) == 0 {
			attName = f.Name
		}
		att, ok := c[attName]
		newContext := fmt.Sprintf("%s.%v", context, f.Name)
		if !ok {
			return &IncompatibleType{context: newContext, to: t, extra: "No attribute with name " + f.Name}
		} else {
			if err := att.Type.CanLoad(f.Type, newContext); err != nil {
				return err
			}
		}
	}
	return nil
}

// Load coerces the given value into a map[string]interface{} where the map values have all been coerced recursively.
// `value` must either be a map with string keys or to a string containing a JSON representation of a map.
// Load also applies any validation rule defined in the composite type attributes.
// Returns `nil` and an error if coercion or validation fails.
func (c Composite) Load(value interface{}) (interface{}, error) {
	// First load from JSON if needed
	var m map[string]interface{}
	switch value.(type) {
	case string:
		if err := json.Unmarshal([]byte(value.(string)), &m); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Composite", extra: "string is not a JSON object"}
		}
	case map[string]interface{}:
		m = value.(map[string]interface{})
	default:
		return nil, &IncompatibleValue{value: value, to: "Composite"}
	}
	if reflect.TypeOf(m).Key().Kind() != reflect.String {
		return nil, &IncompatibleValue{value: value, to: "Composite", extra: "keys must be strings"}
	}
	// Now go through each type member and load and validate value from map
	coerced := make(map[string]interface{})
	errors := make([]error, 0)

	for n, att := range c {
		val, ok := m[n]
		if !ok {
			if att.Required {
				errors = append(errors, &IncompatibleValue{value, "Composite", "missing required attribute " + n})
				continue
			}
			if att.DefaultValue != nil {
				coerced[n] = att.DefaultValue
			}
		} else {
			val, err := att.Type.Load(val)
			if err != nil {
				errors = append(errors, &IncompatibleValue{value, "Composite", fmt.Sprintf("could not load attribute %s: %s", n, err.Error())})
				continue
			}
			allowedValues := att.AllowedValues
			if allowedValues != nil {
				valuesType := reflect.TypeOf(allowedValues).Kind()
				if valuesType != reflect.Slice && valuesType != reflect.Array {
					errors = append(errors, fmt.Errorf("Invalid 'AllowedValues' field, value must be an array but value type is %s", fmt.Sprintf("%s", valuesType)))
					continue
				}
				allowed := reflect.ValueOf(allowedValues)
				ok = (allowed.Len() == 0)
				for i := 0; i < allowed.Len(); i++ {
					if allowed.Index(i).Interface() == val {
						ok = true
						break
					}
				}
				if !ok {
					var values []string
					for i := 0; i < allowed.Len(); i++ {
						values = append(values, fmt.Sprintf("%v", allowed.Index(i).Interface()))
					}
					msg := fmt.Sprintf("value given for attribute %s does not match any of the allowed values (given value was %v, allowed values are %v)",
						n, val, strings.Join(values, ", "))
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			}
			switch att.Type.GetKind() {
			case TString:
				strVal := val.(string)
				if len(att.Regexp) > 0 {
					if ok, _ := regexp.Match(att.Regexp, []byte(strVal)); !ok {
						msg := fmt.Sprintf("value given for attribute %s does not match regular expression %s",
							n, att.Regexp)
						errors = append(errors, &IncompatibleValue{value, "Composite", msg})
						continue
					}
				}
				if len(strVal) < att.MinLength {
					msg := fmt.Sprintf("string value given for attribute %s does not match minimum length restriction (value \"%s\" has less than %v characters)",
						n, strVal, att.MinLength)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
				if att.MaxLength > 0 && len(strVal) > att.MaxLength {
					msg := fmt.Sprintf("string value given for attribute %s does not match maximum length restriction (value \"%s\" has more than %v characters)",
						n, strVal, att.MaxLength)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			case TInteger:
				intVal := val.(int)
				if att.MinValue != nil && intVal < att.MinValue.(int) {
					msg := fmt.Sprintf("integer value given for attribute %s does not match minimum value restriction (value \"%v\" is less than %v)",
						n, intVal, att.MinValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
				if att.MaxValue != nil && intVal > att.MaxValue.(int) {
					msg := fmt.Sprintf("integer value given for attribute %s does not match maximum value restriction (value \"%v\" is more than %v)",
						n, intVal, att.MaxValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			case TFloat:
				floatVal := val.(float64)
				if att.MinValue != nil && floatVal < att.MinValue.(float64) {
					msg := fmt.Sprintf("float value given for attribute %s does not match minimum value restriction (value \"%v\" is less than %v)",
						n, floatVal, att.MinValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
				if att.MaxValue != nil && floatVal > att.MaxValue.(float64) {
					msg := fmt.Sprintf("float value given for attribute %s does not match maximum value restriction (value \"%v\" is more than %v)",
						n, floatVal, att.MaxValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			case TTime:
				timeVal := val.(time.Time)
				if att.MinValue != nil && timeVal.Before(att.MinValue.(time.Time)) {
					msg := fmt.Sprintf("time value given for attribute %s does not match minimum value restriction (value \"%v\" is less than %v)",
						n, timeVal, att.MinValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
				if att.MaxValue != nil && timeVal.After(att.MaxValue.(time.Time)) {
					msg := fmt.Sprintf("time value given for attribute %s does not match maximum value restriction (value \"%v\" is more than %v)",
						n, timeVal, att.MaxValue)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			case TCollection:
				length := reflect.ValueOf(val).Len()
				if length < att.MinLength {
					msg := fmt.Sprintf("collection value given for attribute %s does not match minimum length restriction", n)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
				if att.MaxLength > 0 && length > att.MaxLength {
					msg := fmt.Sprintf("collection value given for attribute %s does not match maximum length restriction", n)
					errors = append(errors, &IncompatibleValue{value, "Composite", msg})
					continue
				}
			}
			coerced[n] = val
		}
	}

	if len(errors) > 0 {
		// TBD create MultiError type
		return nil, errors[0]
	}

	return coerced, nil
}

// GetKind returns the kind of this type (composite)
func (c Composite) GetKind() Kind {
	return TComposite
}

// Collection

// CanLoad checks whether values of the given go type can be loaded into values of this collection type.
// Returns nil if check is successful, error otherwise.
func (c *Collection) CanLoad(t reflect.Type, context string) error {
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return &IncompatibleType{context: context, to: t, extra: "value must be an array or a slice"}
	}
	return c.ElemType.CanLoad(t.Elem(), fmt.Sprintf("%v items", context))
}

// Load coerces the given value into a []interface{} where the array values have all been coerced recursively.
// `value` must either be a slice, an array or a string containing a JSON representation of an array.
// Load also applies any validation rule defined in the collection type element attributes.
// Returns nil and an error if coercion or validation fails.
func (c *Collection) Load(value interface{}) (interface{}, error) {

	var arr []interface{}
	k := reflect.TypeOf(value).Kind()
	if k == reflect.String {
		if err := json.Unmarshal([]byte(value.(string)), &arr); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Collection", extra: fmt.Sprintf("failed to load JSON: %v", err.Error())}
		}
	} else if k == reflect.Slice || k == reflect.Array {
		v := reflect.ValueOf(value)
		for i := 0; i < v.Len(); i++ {
			arr = append(arr, v.Index(i).Interface())
		}
	} else {
		return nil, &IncompatibleValue{value: value, to: "Collection", extra: "value must be an array or a slice"}
	}
	var res []interface{}
	varr := reflect.ValueOf(arr)
	for i := 0; i < varr.Len(); i++ {
		ev, err := c.ElemType.Load(varr.Index(i).Interface())
		if err != nil {
			return nil, &IncompatibleValue{value: value, to: "Collection", extra: fmt.Sprintf("cannot load value at index %v: %v", i, err.Error())}
		}
		res = append(res, ev)
	}
	return interface{}(res), nil
}

// GetKind returns the kind of this type (collection)
func (c *Collection) GetKind() Kind {
	return TCollection
}

// Hash

// CanLoad checks whether values of the given go type can be loaded into values of this hash type.
// Returns nil if check is successful, error otherwise.
func (h *Hash) CanLoad(t reflect.Type, context string) error {
	if t.Kind() != reflect.Map {
		return &IncompatibleType{context: context, to: t, extra: "value must be a map"}
	}
	if t.Key().Kind() != reflect.String {
		return &IncompatibleType{context: context, to: t, extra: "map keys must be strings"}
	}
	return h.ElemType.CanLoad(t.Elem(), fmt.Sprintf("%v hash items", context))
}

// Load coerces the given value into a map[string]interface{} where the map values have all been coerced recursively.
// `value` must either be a map with string keys or a string containing a JSON representation of a map.
// Load also applies any validation rule defined in the hash type element attributes.
// Returns nil and an error if coercion or validation fails.
func (h *Hash) Load(value interface{}) (interface{}, error) {
	var m map[string]interface{}
	k := reflect.TypeOf(value).Kind()
	if k == reflect.String {
		if err := json.Unmarshal([]byte(value.(string)), &m); err != nil {
			return nil, &IncompatibleValue{value: value, to: "Hash", extra: fmt.Sprintf("failed to load JSON: %v", err.Error())}
		}
	} else if k == reflect.Map {
		v := reflect.ValueOf(value)
		keys := v.MapKeys()
		for _, vk := range keys {
			m[vk.String()] = v.MapIndex(vk).Interface()
		}
	} else {
		return nil, &IncompatibleValue{value: value, to: "Hash", extra: "value must be a Hash"}
	}
	var res map[string]interface{}
	vm := reflect.ValueOf(m)
	keys := vm.MapKeys()
	for _, key := range keys {
		ev, err := h.ElemType.Load(vm.MapIndex(key).Interface())
		if err != nil {
			return nil, &IncompatibleValue{value: value, to: "Hash", extra: fmt.Sprintf("cannot load value at key %v: %v", key, err.Error())}
		}
		res[key.String()] = ev
	}
	return interface{}(res), nil
}

// GetKind returns the kind of this type (hash)
func (h *Hash) GetKind() Kind {
	return THash
}
