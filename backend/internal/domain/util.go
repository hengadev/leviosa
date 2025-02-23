package domain

import (
	"fmt"
	"reflect"
)

func CreateModifiedObject[T any](baseObject T, changeMap map[string]any) (*T, error) {
	newObjectPtr := reflect.New(reflect.TypeOf(baseObject)).Interface().(*T)
	v := reflect.ValueOf(newObjectPtr).Elem()
	t := reflect.TypeOf(baseObject)
	vf := reflect.VisibleFields(t)

	*newObjectPtr = baseObject

	for _, field := range vf {
		if value, ok := changeMap[field.Name]; ok {
			fieldValue := v.FieldByName(field.Name)
			switch fieldValue.Kind() {
			// case reflect.Int, reflect.Int64:
			case reflect.Int:
				if val, ok := value.(int); ok {
					fieldValue.SetInt(int64(val))
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected int, got %T", field.Name, value)
				}
			case reflect.String:
				if val, ok := value.(string); ok {
					fieldValue.SetString(val)
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected string, got %T", field.Name, value)
				}
				// ... Add additional cases for other types I want to support
			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Name)
			}
		}
	}
	return newObjectPtr, nil
}

func CreateWithZeroFieldModifiedObject[T any](baseObject T, changeMap map[string]any) (*T, error) {
	newObjectPtr := reflect.New(reflect.TypeOf(baseObject)).Interface().(*T)
	v := reflect.ValueOf(newObjectPtr).Elem()
	t := reflect.TypeOf(baseObject)
	vf := reflect.VisibleFields(t)

	for _, field := range vf {
		if value, ok := changeMap[field.Name]; ok {
			fieldValue := v.FieldByName(field.Name)
			switch fieldValue.Kind() {
			// case reflect.Int, reflect.Int64:
			case reflect.Int:
				if val, ok := value.(int); ok {
					fieldValue.SetInt(int64(val))
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected int, got %T", field.Name, value)
				}
			case reflect.String:
				if val, ok := value.(string); ok {
					fieldValue.SetString(val)
				} else {
					return nil, fmt.Errorf("type mismatch for field %s: expected string, got %T", field.Name, value)
				}
				// ... Add additional cases for other types I want to support
			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Name)
			}
		}
	}
	return newObjectPtr, nil
}
