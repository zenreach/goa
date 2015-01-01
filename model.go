package goa

import (
	"errors"
	"fmt"
	"reflect"
)

// Models contain the REST resource data. They can be instantiated from a REST request payload, from raw database data
// or any other generic representation (JSON or maps keyed by field names).
//
// Model definitions describe the model attributes and a "blueprint" which is the zero value of the go struct used by
// the business logic of the app.
//
// For example the blueprint of a model definition that contains a person name and age attributes could be defined as:
//
//    struct { Name string; Age int }{}
//
// This example would require the model attributes to contain a "Name" attribute of type String and a "Age" attribute
// of type Integer. A blueprint may also use tags to specify the corresponding attribute names:
//
//    type person struct {
//        FirstName string `attribute:"first_name"`
//        LastName  string `attribute:"last_name"`
//    }
//
// Given the "person" struct defined above, the blueprint `person{}` can be used to instantiate a model definition with
// "first_name" and "last_name" attributes:
//
//     personDefinition := NewModel(
//         Attributes{
//             "first_name": Attribute{Type: goa.String, MinLength: 1},
//             "last_name":  Attribute{Type: goa.String, MinLength: 1},
//         },
//         person{}
//     )
//
// Given the definition above the app may instantiate instances of the blueprint type using the `Load()` function.
// The function will take care of validating and coercing the input data into the struct used by the app to implement
// the logic. While model definitions are used internally by the framework to load request payloads they may also be
// used independently by the app for example to load data from a database.
//
// Note that while both model definitions and media types are defined using attributes the semantic is different: model
// definition attributes describe the actual data structure used by the app while media type attributes define how
// responses are built.
type Model struct {
	Attributes     Attributes
	Blueprint      interface{}
	fieldNameByAtt *map[string]string // Internal mapping of blueprint field name to attribute name
}

// Create new model definition given named attributes and a blueprint
// Return an error if the blueprint is invalid (i.e. not a struct) or if the blueprint fields do not match
// the attributes.
func NewModel(attributes Attributes, blueprint interface{}) (*Model, error) {
	bpType := reflect.TypeOf(blueprint)
	if bpType.Kind() != reflect.Struct {
		msg := fmt.Sprintf("Blueprint must be a struct. Given value was a %v.", bpType)
		return nil, NewArgumentError(msg, "blueprint", blueprint)
	}
	// OK we have a valid blueprint type, now let's check the blueprint fields against the attributes
	numField := bpType.NumField()
	if numField != len(attributes) {
		msg := fmt.Sprintf("%v attributes given but blueprint contains %v fields.", len(attributes), numField)
		return nil, NewArgumentError(msg, "blueprint", blueprint)
	}

	for i := 0; i < numField; i++ {
		field := bpType.Field(i)
		attName := attributeName(field)
		if attr, ok := attributes[attName]; !ok {
			msg := fmt.Sprintf("Blueprint field '%s' maps to non-existent attribute '%s'", field.Name, attName)
			return nil, NewArgumentError(msg, "blueprint", blueprint)
		} else if err := attr.Type.CanLoad(field.Type, attName); err != nil {
			msg := fmt.Sprintf("Type of blueprint field '%s' (%s) is incompatible with attribute '%s': %s",
				field.Name, field.Type, attName, err.Error())
			return nil, NewArgumentError(msg, "blueprint", blueprint)
		}
	}

	return &Model{attributes, blueprint, mapFieldNames(bpType, "")}, nil
}

// Load a map indexed by field names or its JSON representation into instance of model definition blueprint struct.
// Argument must be either a string (JSON) or a map whose keys are strings.
// Returns a pointer to struct with the same type as the blueprint whose fields have been initialized from given data.
//
// Example:
//
//    // Data structures used by application logic
//    type Address struct {
//        Street string `attribute:"street"`
//        City   string `attribute:"city"`
//    }
//    type Employee struct {
//        Name    string   `attribute:"name"`
//        Title   string   `attribute:"title"`
//        Address *Address `attribute:"address"`
//    }
//
//    // Model definition attributes
//    attributes := Attributes{
//        "name": Attribute{
//            Type:      String,
//            MinLength: 1,
//            Required:  true,
//        },
//        "title": Attribute{
//            Type:      String,
//            MinLength: 1,
//            Required:  true,
//        },
//        "address": Attribute{
//            Type: Composite{
//                "street": Attribute{
//                    Type: String,
//                },
//                "city": Attribute{
//                    Type:      String,
//                    MinLength: 1,
//                    Required:  true,
//                },
//            },
//        },
//    }
//
//    // Create model definition from attributes and blueprint
//    definition, _ := NewModel(attributes, Employee{})
//
//    // Data coming from external source (API payload, data store etc.)
//    data := map[string]interface{}{
//        "name":  "John",
//        "title": "Accountant",
//        "address": map[string]interface{}{
//            "street": "5779 Lamey Drive",
//            "city":   "Santa Barbara",
//        },
//    }
//
//    // Load data into application data structures
//    if raw, err := definition.Load(&data); err == nil {
//        employee := raw.(*Employee)
//        fmt.Printf("Employee: %+v\n", *employee)
//    } else {
//        fmt.Printf("Load failed: %s\n", err.Error())
//    }
//
func (m *Model) Load(value interface{}) (interface{}, error) {
	val := reflect.New(reflect.TypeOf(m.Blueprint))
	c := Composite(m.Attributes)
	raw, err := c.Load(value)
	if err != nil {
		return nil, err
	}

	rawValue := reflect.ValueOf(raw)
	if err = m.initData(val.Elem(), rawValue, ""); err != nil {
		return nil, err
	}

	return val.Interface(), nil
}

// CanLoad checks whether values of the given go type can be loaded into values
// of this model.
// Returns nil if check is successful, error otherwise.
func (m *Model) CanLoad(t reflect.Type, context string) error {
	c := Composite(m.Attributes)
	return c.CanLoad(t, context)
}

// Validate verifies all model fields recursively.
func (m *Model) Validate() error {
	for n, attr := range m.Attributes {
		if err := attr.Validate(); err != nil {
			return fmt.Errorf("Failed to validate field '%s': %s", n, err.Error())
		}
	}
	if m.Blueprint == nil {
		return errors.New("Model is missing blueprint")
	}
	return nil
}

// Helper method to load data from a map (raw data) into a pointer to struct (blueprint instance)
// This method is recursive, the last argument contains the current "path" to the struct field being init'ed
func (m *Model) initData(data reflect.Value, value reflect.Value, attPrefix string) error {
	for _, k := range value.MapKeys() {
		key := k.String()
		if len(attPrefix) > 0 {
			key = attPrefix + "." + key
		}
		fieldName, _ := (*m.fieldNameByAtt)[key]
		f := data.FieldByName(fieldName)
		if !f.IsValid() {
			return NewErrorf("There is no model attribute named '%s' but argument given to Load() contains a key '%s' with value %v",
				key, key, f.Interface())
		}
		if !f.CanSet() {
			return NewErrorf("Field '%s' cannot be written to, is it public?", fieldName)
		}
		val := value.MapIndex(k).Elem()
		if val.Type().Kind() == reflect.Map {
			if err := m.initData(f, val, key); err != nil {
				return err
			}
		} else {
			if err := m.setFieldValue(f, val, fieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper method used to load given value into given struct field
// Value must have been coerced into a goa supported type
func (m *Model) setFieldValue(field, value reflect.Value, fieldName string) error {
	if err := m.validateFieldKind(field, value.Kind(), fieldName); err != nil {
		return err
	}
	// A coerced value must be one of string, int, float64, bool, time.Time, array or map of values
	switch value.Kind() {
	case reflect.String:
		field.SetString(value.String())
	case reflect.Int:
		i := value.Int()
		if !field.OverflowInt(i) {
			field.SetInt(i)
		}
	case reflect.Float64:
		f := value.Float()
		if !field.OverflowFloat(f) {
			field.SetFloat(f)
		}
	case reflect.Bool:
		field.SetBool(value.Bool())
	case reflect.Array:
		field.Set(reflect.MakeSlice(value.Elem().Type(), value.Len(), value.Len()))
		for i := 0; i < value.Len(); i++ {
			if err := m.setFieldValue(field.Index(i), value.Index(i), fmt.Sprintf("%v[%v]", fieldName, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function used to validate kind of struct field value against attribute type
// Note that this function is called on
func (m *Model) validateFieldKind(field reflect.Value, kind reflect.Kind, name string) error {
	if field.Kind() != kind {
		return NewErrorf("Struct given to Load() defines field '%v' with type %v but the corresponding attribute type is %v", name, field.Kind(), kind)
	}
	return nil
}

// Create map of blueprint struct field names indexed by attribute name
// Used to lookup field name when loading data
func mapFieldNames(blueprint reflect.Type, prefix string) *map[string]string {
	fieldNameByAtt := make(map[string]string)
	for i := 0; i < blueprint.NumField(); i++ {
		field := blueprint.Field(i)
		attName := attributeName(field)
		if len(prefix) > 0 {
			attName = prefix + "." + attName
		}
		fieldNameByAtt[attName] = field.Name
		if field.Type.Kind() == reflect.Struct {
			subMap := mapFieldNames(field.Type, attName)
			for k, v := range *subMap {
				fieldNameByAtt[k] = v
			}
		}
	}

	return &fieldNameByAtt
}

// Retrieve name of attribute that corresponds to blueprint struct field
func attributeName(field reflect.StructField) string {
	attName := field.Tag.Get("attribute")
	if len(attName) == 0 {
		attName = field.Name
	}
	return attName
}
