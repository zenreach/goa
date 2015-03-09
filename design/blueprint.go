package design

import (
	"fmt"
	"reflect"
)

// A blueprint consists of a struct and an Object describing the struct fields.
// Serialized representations of a blueprint can be validated against the object properties and
// loaded into an instance of its type.
type Blueprint struct {
	Properties  Object
	Type        interface{}
	fieldByProp map[string]string // Internal mapping of type field name to property name
}

// NewBlueprint validates that the given Object describes the given type and creates a blueprint
// or returns an error accordingly.
func NewBlueprint(object Object, typ interface{}) (*Blueprint, error) {
	bpType := reflect.TypeOf(typ)
	if bpType == nil || bpType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("Type of blueprint must be a struct. Given value was a %v.", bpType)
	}
	// OK we have a valid type, now let's check its fields against the object properties
	numField := bpType.NumField()
	if numField != len(object) {
		return nil, fmt.Errorf("Object contains %d property(ies) but type contains %d field(s).", len(object), numField)
	}

	for i := 0; i < numField; i++ {
		field := bpType.Field(i)
		name := propertyName(field)
		if prop, ok := object[name]; !ok {
			return nil, fmt.Errorf("Type field '%s' maps to non-existent property '%s'", field.Name, name)
		} else if err := prop.Type.CanLoad(field.Type, name); err != nil {
			return nil, fmt.Errorf("Type of field '%s' (%s) is incompatible with property '%s': %s",
				field.Name, field.Type, name, err.Error())
		}
	}
	b := Blueprint{Properties: object, Type: typ, fieldByProp: mapFieldNames(bpType, "")}
	return &b, nil
}

// Type name
// Blueprint implements DataType so blueprints can be used everywhere types can.
func (b *Blueprint) Name() string {
	return "object"
}

// Type kind
// Blueprint implements DataType so blueprints can be used everywhere types can.
func (b *Blueprint) Kind() Kind {
	return ObjectType
}

// Load a map indexed by field names or its JSON representation into instance of blueprint type struct.
// Argument must be either a string (JSON) or a map whose keys are strings.
// Returns a pointer to struct with the same type as the blueprint whose fields have been initialized from given data.
func (b *Blueprint) Load(value interface{}) (interface{}, error) {
	val := reflect.New(reflect.TypeOf(b.Type))
	raw, err := b.Properties.Load(value)
	if err != nil {
		return nil, err
	}

	rawValue := reflect.ValueOf(raw)
	if err = b.initData(val.Elem(), rawValue, ""); err != nil {
		return nil, err
	}

	return val.Interface(), nil
}

// CanLoad checks whether values of the given go type can be loaded into values of this blueprint.
// Returns nil if check is successful, error otherwise.
func (b *Blueprint) CanLoad(t reflect.Type, context string) error {
	return b.Properties.CanLoad(t, context)
}

// Helper method to load data from a map (raw data) into a pointer to struct.
// This method is recursive, the last argument contains the current "path" to the struct field being init'ed.
func (b *Blueprint) initData(data reflect.Value, value reflect.Value, attPrefix string) error {
	for _, k := range value.MapKeys() {
		key := k.String()
		if len(attPrefix) > 0 {
			key = attPrefix + "." + key
		}
		fieldName, _ := b.fieldByProp[key]
		f := data.FieldByName(fieldName)
		if !f.IsValid() {
			return fmt.Errorf("There is no model attribute named '%s' but argument given to Load() contains a key '%s' with value '%v'",
				key, key, value.MapIndex(k).Interface())
		}
		if !f.CanSet() {
			return fmt.Errorf("Field '%s' cannot be written to, is it public?", fieldName)
		}
		val := value.MapIndex(k).Elem()
		if val.Type().Kind() == reflect.Map {
			if err := b.initData(f, val, key); err != nil {
				return err
			}
		} else {
			if err := b.setFieldValue(f, val, fieldName); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper method used to load given value into given struct field
// Value must have been coerced into a goa supported type
func (b *Blueprint) setFieldValue(field, value reflect.Value, fieldName string) error {
	if err := b.validateFieldKind(field, value.Kind(), fieldName); err != nil {
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
			if err := b.setFieldValue(field.Index(i), value.Index(i), fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper function used to validate kind of struct field value against attribute type
func (b *Blueprint) validateFieldKind(field reflect.Value, kind reflect.Kind, name string) error {
	if field.Kind() != kind {
		return fmt.Errorf("Struct given to Load() defines field '%s' with type %v but the corresponding attribute type is %v",
			name, field.Kind(), kind)
	}
	return nil
}

// Compute name of property that corresponds to type struct field.
// Check if struct field has a "property" tag and if so use that, otherwise use field name.
func propertyName(field reflect.StructField) string {
	name := field.Tag.Get("property")
	if len(name) == 0 {
		name = field.Name
	}
	return name
}

// Create map of struct field names indexed by property name.
// Used to lookup field name when loading data.
func mapFieldNames(typ reflect.Type, prefix string) map[string]string {
	fieldNameByProp := make(map[string]string)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := propertyName(field)
		if len(prefix) > 0 {
			name = prefix + "." + name
		}
		fieldNameByProp[name] = field.Name
		if field.Type.Kind() == reflect.Struct {
			subMap := mapFieldNames(field.Type, name)
			for k, v := range subMap {
				fieldNameByProp[k] = v
			}
		}
	}
	return fieldNameByProp
}
