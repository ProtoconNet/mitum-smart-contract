package runtime

import (
	"fmt"
	"strconv"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type QueryValue struct {
	Kind    string          `json:"kind"`
	Scalar  string          `json:"scalar,omitempty"`
	Fields  []QueryField    `json:"fields,omitempty"`
	Entries []QueryMapEntry `json:"entries,omitempty"`
	Items   []QueryValue    `json:"items,omitempty"`
	IsNil   bool            `json:"is_nil,omitempty"`
}

type QueryField struct {
	Name  string     `json:"name"`
	Value QueryValue `json:"value"`
}

type QueryMapEntry struct {
	Key   string     `json:"key"`
	Value QueryValue `json:"value"`
}

func ExtractQueryValue(schema ContractSchema, typ TypeRef, tv gno.TypedValue, store gno.Store) (QueryValue, error) {
	tvCopy := tv
	deepFillSnapshotTypedValue(schema, typ, &tvCopy, store)

	sv, err := ExtractSnapshotValue(schema, typ, tvCopy)
	if err != nil {
		return QueryValue{}, err
	}

	return snapshotValueToQueryValue(sv), nil
}

func QueryValueToJSONCompatible(schema ContractSchema, typ TypeRef, qv QueryValue) (interface{}, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		return queryScalarToJSONCompatible(resolved, qv)

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return nil, fmt.Errorf("unsupported query result type %q; %s", typ.String(), QueryResultSupportDescription)
		}

		if qv.Kind != string(TypeStruct) {
			return nil, fmt.Errorf("expected struct query value for type %q", typ.String())
		}
		if len(qv.Fields) != len(resolved.Fields) {
			return nil, fmt.Errorf("expected %d struct fields for %q, got %d", len(resolved.Fields), typ.String(), len(qv.Fields))
		}

		out := make(map[string]interface{}, len(resolved.Fields))
		for i, field := range resolved.Fields {
			qf := qv.Fields[i]
			if qf.Name != field.Name {
				return nil, fmt.Errorf("expected struct field %q at index %d for %q, got %q", field.Name, i, typ.String(), qf.Name)
			}

			v, err := QueryValueToJSONCompatible(schema, field.Type, qf.Value)
			if err != nil {
				return nil, err
			}
			out[field.Name] = v
		}

		return out, nil

	case TypeMap:
		if qv.Kind != string(TypeMap) {
			return nil, fmt.Errorf("expected map query value for type %q", typ.String())
		}
		if resolved.Key == nil || resolved.Elem == nil {
			return nil, fmt.Errorf("unsupported query result type %q; %s", typ.String(), QueryResultSupportDescription)
		}
		if qv.IsNil {
			return nil, nil
		}

		out := make(map[string]interface{}, len(qv.Entries))
		for _, entry := range qv.Entries {
			v, err := QueryValueToJSONCompatible(schema, *resolved.Elem, entry.Value)
			if err != nil {
				return nil, err
			}
			out[entry.Key] = v
		}
		return out, nil

	case TypeSlice:
		if qv.Kind != string(TypeSlice) {
			return nil, fmt.Errorf("expected slice query value for type %q", typ.String())
		}
		if resolved.Elem == nil {
			return nil, fmt.Errorf("unsupported query result type %q; %s", typ.String(), QueryResultSupportDescription)
		}
		if qv.IsNil {
			return nil, nil
		}

		out := make([]interface{}, 0, len(qv.Items))
		for _, item := range qv.Items {
			v, err := QueryValueToJSONCompatible(schema, *resolved.Elem, item)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil

	default:
		return nil, fmt.Errorf("unsupported query result type %q; %s", typ.String(), QueryResultSupportDescription)
	}
}

func snapshotValueToQueryValue(sv SnapshotValue) QueryValue {
	out := QueryValue{
		Kind:   sv.Kind,
		Scalar: sv.Scalar,
		IsNil:  sv.IsNil,
	}

	if len(sv.Fields) > 0 {
		out.Fields = make([]QueryField, 0, len(sv.Fields))
		for _, field := range sv.Fields {
			out.Fields = append(out.Fields, QueryField{
				Name:  field.Name,
				Value: snapshotValueToQueryValue(field.Value),
			})
		}
	}

	if len(sv.Entries) > 0 {
		out.Entries = make([]QueryMapEntry, 0, len(sv.Entries))
		for _, entry := range sv.Entries {
			out.Entries = append(out.Entries, QueryMapEntry{
				Key:   entry.Key,
				Value: snapshotValueToQueryValue(entry.Value),
			})
		}
	}

	if len(sv.Items) > 0 {
		out.Items = make([]QueryValue, 0, len(sv.Items))
		for _, item := range sv.Items {
			out.Items = append(out.Items, snapshotValueToQueryValue(item))
		}
	}

	return out
}

func queryScalarToJSONCompatible(typ TypeRef, qv QueryValue) (interface{}, error) {
	if qv.Kind != string(TypeScalar) {
		return nil, fmt.Errorf("expected scalar query value for type %q", typ.String())
	}

	switch typ.NormalizedString() {
	case "string":
		return qv.Scalar, nil
	case "bool":
		switch qv.Scalar {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool query value %q", qv.Scalar)
		}
	case "int":
		v, err := strconv.Atoi(qv.Scalar)
		if err != nil {
			return nil, fmt.Errorf("invalid int query value %q: %w", qv.Scalar, err)
		}
		return v, nil
	case "int64":
		v, err := strconv.ParseInt(qv.Scalar, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int query value %q: %w", qv.Scalar, err)
		}
		return v, nil
	case "uint64":
		v, err := strconv.ParseUint(qv.Scalar, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid uint64 query value %q: %w", qv.Scalar, err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported query result type %q; %s", typ.String(), QueryResultSupportDescription)
	}
}
