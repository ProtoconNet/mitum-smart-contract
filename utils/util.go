package utils

import (
	"github.com/pkg/errors"
	"reflect"
)

func HasValue(slice interface{}, value interface{}) (bool, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false, errors.Errorf("Invalid data-type, Not Slice")
	}

	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == value {
			return true, nil
		}
	}

	return false, nil
}

func MustHasValue(slice interface{}, value interface{}) (interface{}, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, errors.Errorf("Invalid data-type, Not Slice")
	}

	var nSlice reflect.Value
	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == value {
			nSlice = reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
			reflect.Copy(nSlice, v)
			break
		}
		if i == v.Len()-1 {
			nSlice = reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
			reflect.Copy(nSlice, v)
			nSlice = reflect.Append(nSlice, reflect.ValueOf(value))
		}
	}

	return nSlice.Interface(), nil
}

func HasFieldValue(slice interface{}, fieldName string, value interface{}) (bool, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false, errors.Errorf("Invalid data-type, Not Slice")
	}

	if len(fieldName) < 1 {
		return false, errors.Errorf("empty field name")
	}

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Struct {
			field := item.FieldByName(fieldName)
			if !field.IsValid() {
				panic("No such field: " + fieldName)
			}
			if field.Interface() == value {
				return true, nil
			}
		}
	}

	return false, nil
}

func HasFieldAndSliceValue(
	slice interface{},
	fieldName, sliceFieldName string,
	fieldValue, sliceFieldValue interface{},
) (bool, bool, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false, false, errors.Errorf("Invalid data-type, Not Slice")
	}

	if len(fieldName) < 1 {
		return false, false, errors.Errorf("empty field name")
	} else if len(sliceFieldName) < 1 {
		return false, false, errors.Errorf("empty field name for slice")
	}

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() != reflect.Struct {
			return false, false, errors.Errorf("Invalid data-type, Not Struct")
		}
		field := item.FieldByName(fieldName)
		if !field.IsValid() {
			return false, false, errors.Errorf("No such field: " + fieldName)
		}
		if field.Interface() != fieldValue {
			return false, false, nil
		}
		sliceField := item.FieldByName(sliceFieldName)
		if !sliceField.IsValid() {
			return true, false, errors.Errorf("No such field: " + sliceFieldName)
		}
		if sliceField.Kind() != reflect.Slice {
			return true, false, errors.Errorf("Invalid data-type, Not Slice")
		}
		for j := 0; j < sliceField.Len(); j++ {
			if sliceField.Index(j).Interface() == sliceFieldValue {
				return true, true, nil
			}
			if i == sliceField.Len()-1 {
				return true, false, nil
			}
		}
	}
	return false, false, nil
}
