package utils

import "reflect"

func ContainsValue(slice interface{}, fieldName string, value interface{}) bool {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Struct {
			field := item.FieldByName(fieldName)
			if !field.IsValid() {
				panic("No such field: " + fieldName)
			}
			if field.Interface() == value {
				return true
			}
		}
	}

	return false
}
