package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func buildWriteCallArgExpr(schema ContractSchema, typ TypeRef, raw string) (gno.Expr, error) {
	return buildCallArgExpr(schema, typ, raw, WriteArgSupportDescription)
}

func buildQueryCallArgExpr(schema ContractSchema, typ TypeRef, raw string) (gno.Expr, error) {
	return buildCallArgExpr(schema, typ, raw, QueryArgSupportDescription)
}

func buildCallArgExpr(schema ContractSchema, typ TypeRef, raw string, description string) (gno.Expr, error) {
	resolved := schema.ResolveType(typ)
	if resolved.IsScalar() {
		return scalarArgExpr(resolved, raw)
	}

	value, err := DecodeCallArgValue(schema, typ, raw, description)
	if err != nil {
		return nil, err
	}

	return BuildExpr(schema, typ, value)
}

func DecodeCallArgValue(schema ContractSchema, typ TypeRef, raw string, description string) (SnapshotValue, error) {
	resolved := schema.ResolveType(typ)
	if resolved.IsScalar() {
		return decodeScalarCallArgValue(resolved, raw)
	}

	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()

	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return SnapshotValue{}, fmt.Errorf("expected JSON value for %q: %w", typ.String(), err)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return SnapshotValue{}, fmt.Errorf("expected single JSON value for %q", typ.String())
		}
		return SnapshotValue{}, fmt.Errorf("invalid trailing JSON for %q: %w", typ.String(), err)
	}

	return decodeCompositeCallArgValue(schema, typ, data, fmt.Sprintf("arg %q", typ.String()), description)
}

func decodeCompositeCallArgValue(schema ContractSchema, typ TypeRef, data interface{}, subject string, description string) (SnapshotValue, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		return decodeScalarJSONValue(resolved, data, subject)

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return SnapshotValue{}, fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
		}
		return decodeStructCallArgValue(schema, typ, resolved, data, subject, description)

	case TypeMap:
		return decodeMapCallArgValue(schema, typ, resolved, data, subject, description)

	case TypeSlice:
		return decodeSliceCallArgValue(schema, typ, resolved, data, subject, description)

	default:
		return SnapshotValue{}, fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
	}
}

func decodeStructCallArgValue(schema ContractSchema, typ TypeRef, resolved TypeRef, data interface{}, subject string, description string) (SnapshotValue, error) {
	if _, err := schema.validateNamedStructType(typ, subject); err != nil {
		return SnapshotValue{}, fmt.Errorf("%w; %s", err, description)
	}

	obj, ok := data.(map[string]interface{})
	if !ok {
		return SnapshotValue{}, fmt.Errorf("%s expects JSON object for type %q", subject, typ.String())
	}

	fields := make([]SnapshotField, 0, len(resolved.Fields))
	usedKeys := map[string]bool{}
	for _, field := range resolved.Fields {
		rawValue, found, usedKey, err := lookupJSONObjectField(obj, field.Name)
		if err != nil {
			return SnapshotValue{}, fmt.Errorf("%s field %q: %w", subject, field.Name, err)
		}
		var value SnapshotValue
		if found {
			usedKeys[usedKey] = true
			value, err = decodeCompositeCallArgValue(schema, field.Type, rawValue, fmt.Sprintf("%s field %q", subject, field.Name), description)
			if err != nil {
				return SnapshotValue{}, err
			}
		} else {
			value, err = zeroValueForType(schema, field.Type, description)
			if err != nil {
				return SnapshotValue{}, err
			}
		}

		fields = append(fields, SnapshotField{Name: field.Name, Value: value})
	}

	for key := range obj {
		if !usedKeys[key] {
			return SnapshotValue{}, fmt.Errorf("%s contains unknown field %q", subject, key)
		}
	}

	return SnapshotValue{
		Kind:   string(TypeStruct),
		Fields: fields,
	}, nil
}

func decodeMapCallArgValue(schema ContractSchema, typ TypeRef, resolved TypeRef, data interface{}, subject string, description string) (SnapshotValue, error) {
	if err := schema.validateMapTypeRecursive(typ, subject, map[string]bool{}); err != nil {
		return SnapshotValue{}, fmt.Errorf("%w; %s", err, description)
	}
	if resolved.Key == nil || resolved.Elem == nil {
		return SnapshotValue{}, fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
	}

	if data == nil {
		return SnapshotValue{Kind: string(TypeMap), IsNil: true}, nil
	}

	obj, ok := data.(map[string]interface{})
	if !ok {
		return SnapshotValue{}, fmt.Errorf("%s expects JSON object for type %q", subject, typ.String())
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]SnapshotMapEntry, 0, len(keys))
	for _, key := range keys {
		value, err := decodeCompositeCallArgValue(schema, *resolved.Elem, obj[key], fmt.Sprintf("%s[%q]", subject, key), description)
		if err != nil {
			return SnapshotValue{}, err
		}
		entries = append(entries, SnapshotMapEntry{Key: key, Value: value})
	}

	return SnapshotValue{
		Kind:    string(TypeMap),
		Entries: entries,
	}, nil
}

func decodeSliceCallArgValue(schema ContractSchema, typ TypeRef, resolved TypeRef, data interface{}, subject string, description string) (SnapshotValue, error) {
	if err := schema.validateSliceTypeRecursive(typ, subject, map[string]bool{}); err != nil {
		return SnapshotValue{}, fmt.Errorf("%w; %s", err, description)
	}
	if resolved.Elem == nil {
		return SnapshotValue{}, fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
	}

	if data == nil {
		return SnapshotValue{Kind: string(TypeSlice), IsNil: true}, nil
	}

	arr, ok := data.([]interface{})
	if !ok {
		return SnapshotValue{}, fmt.Errorf("%s expects JSON array for type %q", subject, typ.String())
	}

	items := make([]SnapshotValue, 0, len(arr))
	for i, item := range arr {
		value, err := decodeCompositeCallArgValue(schema, *resolved.Elem, item, fmt.Sprintf("%s[%d]", subject, i), description)
		if err != nil {
			return SnapshotValue{}, err
		}
		items = append(items, value)
	}

	return SnapshotValue{
		Kind:  string(TypeSlice),
		Items: items,
	}, nil
}

func decodeScalarCallArgValue(typ TypeRef, raw string) (SnapshotValue, error) {
	switch typ.NormalizedString() {
	case "string":
		return SnapshotValue{Kind: string(TypeScalar), Scalar: raw}, nil
	case "bool":
		if raw != "true" && raw != "false" {
			return SnapshotValue{}, fmt.Errorf("expected bool string")
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: raw}, nil
	case "int", "int64":
		if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
			return SnapshotValue{}, err
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: raw}, nil
	case "uint64":
		if _, err := strconv.ParseUint(raw, 10, 64); err != nil {
			return SnapshotValue{}, err
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: raw}, nil
	default:
		return SnapshotValue{}, fmt.Errorf("unsupported arg type %q; %s", typ.String(), ScalarOnlySupportDescription)
	}
}

func decodeScalarJSONValue(typ TypeRef, data interface{}, subject string) (SnapshotValue, error) {
	switch typ.NormalizedString() {
	case "string":
		s, ok := data.(string)
		if !ok {
			return SnapshotValue{}, fmt.Errorf("%s expects JSON string", subject)
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: s}, nil
	case "bool":
		b, ok := data.(bool)
		if !ok {
			return SnapshotValue{}, fmt.Errorf("%s expects JSON bool", subject)
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: strconv.FormatBool(b)}, nil
	case "int":
		n, err := decodeJSONInteger(data, subject)
		if err != nil {
			return SnapshotValue{}, err
		}
		if strconv.IntSize == 32 && (n < -2147483648 || n > 2147483647) {
			return SnapshotValue{}, fmt.Errorf("%s integer %d overflows int", subject, n)
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: strconv.FormatInt(n, 10)}, nil
	case "int64":
		n, err := decodeJSONInteger(data, subject)
		if err != nil {
			return SnapshotValue{}, err
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: strconv.FormatInt(n, 10)}, nil
	case "uint64":
		u, err := decodeJSONUnsignedInteger(data, subject)
		if err != nil {
			return SnapshotValue{}, err
		}
		return SnapshotValue{Kind: string(TypeScalar), Scalar: strconv.FormatUint(u, 10)}, nil
	default:
		return SnapshotValue{}, fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), WriteArgSupportDescription)
	}
}

func zeroValueForType(schema ContractSchema, typ TypeRef, description string) (SnapshotValue, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		switch resolved.NormalizedString() {
		case "string":
			return SnapshotValue{Kind: string(TypeScalar), Scalar: ""}, nil
		case "bool":
			return SnapshotValue{Kind: string(TypeScalar), Scalar: "false"}, nil
		case "int", "int64", "uint64":
			return SnapshotValue{Kind: string(TypeScalar), Scalar: "0"}, nil
		default:
			return SnapshotValue{}, fmt.Errorf("unsupported zero value type %q; %s", typ.String(), description)
		}

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return SnapshotValue{}, fmt.Errorf("unsupported zero value type %q; %s", typ.String(), description)
		}
		fields := make([]SnapshotField, 0, len(resolved.Fields))
		for _, field := range resolved.Fields {
			value, err := zeroValueForType(schema, field.Type, description)
			if err != nil {
				return SnapshotValue{}, err
			}
			fields = append(fields, SnapshotField{Name: field.Name, Value: value})
		}
		return SnapshotValue{Kind: string(TypeStruct), Fields: fields}, nil

	case TypeMap:
		return SnapshotValue{Kind: string(TypeMap), IsNil: true}, nil

	case TypeSlice:
		return SnapshotValue{Kind: string(TypeSlice), IsNil: true}, nil

	default:
		return SnapshotValue{}, fmt.Errorf("unsupported zero value type %q; %s", typ.String(), description)
	}
}

func lookupJSONObjectField(obj map[string]interface{}, fieldName string) (interface{}, bool, string, error) {
	if value, found := obj[fieldName]; found {
		return value, true, fieldName, nil
	}

	matchedKey := ""
	var matchedValue interface{}
	for key, value := range obj {
		if strings.EqualFold(key, fieldName) {
			if matchedKey != "" && matchedKey != key {
				return nil, false, "", fmt.Errorf("ambiguous field match for %q", fieldName)
			}
			matchedKey = key
			matchedValue = value
		}
	}
	if matchedKey != "" {
		return matchedValue, true, matchedKey, nil
	}

	return nil, false, "", nil
}

func scalarArgExpr(typ TypeRef, raw string) (gno.Expr, error) {
	switch typ.NormalizedString() {
	case "string":
		return gno.Str(raw), nil
	case "bool":
		if raw != "true" && raw != "false" {
			return nil, fmt.Errorf("expected bool string")
		}
		return gno.X(raw), nil
	case "int", "int64":
		if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
			return nil, err
		}
		return gno.Num(raw), nil
	case "uint64":
		if _, err := strconv.ParseUint(raw, 10, 64); err != nil {
			return nil, err
		}
		return gno.Num(raw), nil
	default:
		return nil, fmt.Errorf("unsupported arg type %q; %s", typ.String(), ScalarOnlySupportDescription)
	}
}

func decodeJSONInteger(data interface{}, subject string) (int64, error) {
	switch v := data.(type) {
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s expects JSON integer: %w", subject, err)
		}
		return n, nil
	case float64:
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("%s expects JSON integer", subject)
		}
		return int64(v), nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("%s expects JSON integer", subject)
	}
}

func decodeJSONUnsignedInteger(data interface{}, subject string) (uint64, error) {
	switch v := data.(type) {
	case json.Number:
		n, err := strconv.ParseUint(v.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s expects JSON unsigned integer: %w", subject, err)
		}
		return n, nil
	case float64:
		if v != float64(uint64(v)) {
			return 0, fmt.Errorf("%s expects JSON unsigned integer", subject)
		}
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("%s expects JSON unsigned integer", subject)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("%s expects JSON unsigned integer", subject)
		}
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("%s expects JSON unsigned integer", subject)
	}
}
