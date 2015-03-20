package goa

import (
	"fmt"
	"reflect"
)

// InitStruct loads data from a map into a struct recursively.
func InitStruct(inited interface{}, data map[string]interface{}) error {
	initType := reflect.TypeOf(inited)
	if initType == nil || initType.Kind() != reflect.Ptr {
		return fmt.Errorf("invalid inited value, must be a pointer - got %v", initType)
	}
	sType := initType.Elem()
	if sType == nil || sType.Kind() != reflect.Struct {
		return fmt.Errorf("invalid inited value, must be a pointer on struct - got pointer on %v", sType)
	}
	value := reflect.Zero(sType)
	if err := initData(reflect.ValueOf(value), reflect.ValueOf(data), ""); err != nil {
		return err
	}
	reflect.ValueOf(inited).Elem().Set(value)
	return nil
}

// Initialize data structure recursively using provided data (map of string to interface).
// Last argument is path to field currently being init'ed (using dot notation).
func initData(value reflect.Value, data reflect.Value, attPrefix string) error {
	for _, k := range data.MapKeys() {
		key := k.String()
		if len(attPrefix) > 0 {
			key = attPrefix + "." + key
		}
		f := value.FieldByName(key)
		if !f.IsValid() {
			return fmt.Errorf("unknown %v field '%s'", value.Type(), key)
		}
		if !f.CanSet() {
			return fmt.Errorf("%v field '%s' cannot be written to, is it public?",
				value.Type(), key)
		}
		val := data.MapIndex(k).Elem()
		if val.Type().Kind() == reflect.Map {
			if err := initData(f, val, key); err != nil {
				return err
			}
		} else {
			if err := setFieldValue(f, val, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// setFieldValue loads given value into given struct field.
// Value type must be a JSON schema primitive type.
func setFieldValue(field, value reflect.Value, fieldName string) error {
	if err := validateFieldKind(field, value.Kind(), fieldName); err != nil {
		return fmt.Errorf("field '%s': %s", fieldName, err)
	}
	// value must be a string, int, float64, bool, array or map of values
	switch value.Kind() {
	case reflect.String:
		field.SetString(value.String())
	case reflect.Int:
		i := value.Int()
		if !field.OverflowInt(i) {
			field.SetInt(i)
		} else {
			return fmt.Errorf("field '%s': integer value too big", fieldName)
		}
	case reflect.Float64:
		f := value.Float()
		if !field.OverflowFloat(f) {
			field.SetFloat(f)
		} else {
			return fmt.Errorf("field '%s': float value too big", fieldName)
		}
	case reflect.Bool:
		field.SetBool(value.Bool())
	case reflect.Array:
		field.Set(reflect.MakeSlice(value.Elem().Type(), value.Len(), value.Len()))
		for i := 0; i < value.Len(); i++ {
			if err := setFieldValue(field.Index(i), value.Index(i),
				fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
				return fmt.Errorf("field '%s' item %d: %s", fieldName, i, err)
			}
		}
	}

	return nil
}

// Helper function used to validate kind of struct field value against attribute type
func validateFieldKind(field reflect.Value, kind reflect.Kind, name string) error {
	if field.Kind() != kind {
		return fmt.Errorf("invalid value type '%v'", kind)
	}
	return nil
}
