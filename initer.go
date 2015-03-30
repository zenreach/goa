package goa

import (
	"fmt"
	"reflect"
	"time"
)

// InitStruct loads data from a map into a struct recursively.
func InitStruct(inited interface{}, data map[string]interface{}) error {
	initVal := reflect.ValueOf(inited)
	if initVal.Kind() != reflect.Ptr {
		return fmt.Errorf("invalid inited value, must be a pointer - got %s", initVal.Type().Name())
	}
	sVal := initVal.Elem()
	if sVal.Kind() != reflect.Struct {
		return fmt.Errorf("invalid inited value, must be a pointer on struct - got pointer on %s", sVal.Type().Name())
	}
	if err := initData(sVal, reflect.ValueOf(data)); err != nil {
		return err
	}
	return nil
}

// Initialize data structure recursively using provided data (map of string to interface).
func initData(value reflect.Value, data reflect.Value) error {
	for _, k := range data.MapKeys() {
		val := data.MapIndex(k).Interface()
		if val == nil {
			// This is OK, maybe the payload object defines less fields than
			// the object used to load the data but the extra fields are all nil.
			continue
		}
		key := k.String()
		f := value.FieldByName(key)
		if !f.IsValid() {
			return fmt.Errorf("unknown %v field '%s'", value.Type().Name(), key)
		}
		if !f.CanSet() {
			return fmt.Errorf("%v field '%s' cannot be written to, is it public?",
				value.Type().Name(), key)
		}
		d := reflect.ValueOf(val)
		if d.Kind() == reflect.Invalid || d.Interface() == nil {
			// No value for that field
			continue
		}
		if d.Kind() == reflect.Map {
			if f.Kind() != reflect.Ptr {
				return fmt.Errorf("invalid field %v, must be a struct pointer but is a %s", key, f.Kind())
			}
			f = f.Elem()
			if err := initData(f, d); err != nil {
				return err
			}
		} else {
			if err := setFieldValue(f, d, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// setFieldValue loads given value into given struct field.
// Value type must be a JSON schema primitive type.
func setFieldValue(field, value reflect.Value, fieldName string) error {
	// value must be a string, int, float64, bool, array or map of values
	switch value.Kind() {
	case reflect.String:
		if _, ok := field.Interface().(time.Time); ok {
			// Make a special case for time.Time struct fields as this is a common
			// occurrence not supported natively by JSON.
			t, err := time.Parse(time.RFC3339, value.Interface().(string))
			if err != nil {
				return fmt.Errorf("field '%s': invalid time value %v",
					fieldName, value.Interface())
			}
			field.Set(reflect.ValueOf(t))
		} else {
			field.SetString(value.String())
		}
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
	default:
		return fmt.Errorf("unsupported data type %s for field '%s'", value.Kind(), fieldName)
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
