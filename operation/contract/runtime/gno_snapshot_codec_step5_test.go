package runtime

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestExtractSnapshotValueMapStringStructDeterministicOrder(t *testing.T) {
	schema := userMapSchema()
	typ := mapStringUserType()

	valueA, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: userStructTypedValue(20, false)},
		mapTestEntry{key: "alice", value: userStructTypedValue(10, true)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	valueB, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "alice", value: userStructTypedValue(10, true)},
		mapTestEntry{key: "bob", value: userStructTypedValue(20, false)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	bytesA := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "users", Value: valueA}},
	})
	bytesB := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "users", Value: valueB}},
	})

	if string(bytesA) != string(bytesB) {
		t.Fatalf("expected deterministic bytes for same logical map state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	if len(valueA.Entries) != 2 || valueA.Entries[0].Key != "alice" || valueA.Entries[1].Key != "bob" {
		t.Fatalf("expected lexical key order, got %#v", valueA.Entries)
	}
}

func TestBuildLiteralMapStringStructRoundTrip(t *testing.T) {
	schema := userMapSchema()
	typ := mapStringUserType()

	value, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: userStructTypedValue(20, false)},
		mapTestEntry{key: "alice", value: userStructTypedValue(10, true)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}

	if lit != `map[string]User{"alice":User{Balance:10,Active:true},"bob":User{Balance:20,Active:false}}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestExtractSnapshotValueMapStringStructNilAndEmptyMapPolicy(t *testing.T) {
	schema := userMapSchema()
	typ := mapStringUserType()

	nilValue, err := ExtractSnapshotValue(schema, typ, gnoTypedNilMap())
	if err != nil {
		t.Fatalf("ExtractSnapshotValue(nil) returned error: %v", err)
	}
	if !nilValue.IsNil || len(nilValue.Entries) != 0 {
		t.Fatalf("expected nil map snapshot, got %#v", nilValue)
	}

	nilLit, err := BuildLiteral(schema, typ, nilValue)
	if err != nil {
		t.Fatalf("BuildLiteral(nil map) returned error: %v", err)
	}
	if nilLit != "nil" {
		t.Fatalf("unexpected nil map literal: %s", nilLit)
	}

	emptyValue, err := ExtractSnapshotValue(schema, typ, emptyMapTypedValue())
	if err != nil {
		t.Fatalf("ExtractSnapshotValue(empty) returned error: %v", err)
	}
	if emptyValue.IsNil || len(emptyValue.Entries) != 0 {
		t.Fatalf("expected empty non-nil map snapshot, got %#v", emptyValue)
	}

	emptyLit, err := BuildLiteral(schema, typ, emptyValue)
	if err != nil {
		t.Fatalf("BuildLiteral(empty map) returned error: %v", err)
	}
	if emptyLit != `map[string]User{}` {
		t.Fatalf("unexpected empty map literal: %s", emptyLit)
	}
}

func TestExtractSnapshotValueMapStringStructWithNestedStructUnsupported(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Meta": {
					Kind: TypeStruct,
					Name: "Meta",
					Raw:  "Meta",
					Fields: []StructField{
						{Name: "Count", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
					},
				},
				"User": {
					Kind: TypeStruct,
					Name: "User",
					Raw:  "User",
					Fields: []StructField{
						{Name: "Meta", Type: TypeRef{Kind: TypeNamed, Raw: "Meta", Name: "Meta"}},
					},
				},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, mapStringNamedType("User"), emptyMapTypedValue())
	if err == nil {
		t.Fatalf("expected nested struct map value unsupported error")
	}
}

func TestExtractSnapshotValueMapStringStructWithMapFieldUnsupported(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"User": {
					Kind: TypeStruct,
					Name: "User",
					Raw:  "User",
					Fields: []StructField{
						{Name: "Flags", Type: TypeRef{
							Kind: TypeMap,
							Raw:  "map[string]bool",
							Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
							Elem: &TypeRef{Kind: TypeScalar, Raw: "bool", Scalar: "bool"},
						}},
					},
				},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, mapStringNamedType("User"), emptyMapTypedValue())
	if err == nil {
		t.Fatalf("expected map field inside map value struct unsupported error")
	}
}

func TestExtractSnapshotValueMapStringStructWithSliceFieldUnsupported(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"User": {
					Kind: TypeStruct,
					Name: "User",
					Raw:  "User",
					Fields: []StructField{
						{Name: "Tags", Type: TypeRef{
							Kind: TypeSlice,
							Raw:  "[]string",
							Elem: &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
						}},
					},
				},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, mapStringNamedType("User"), emptyMapTypedValue())
	if err == nil {
		t.Fatalf("expected slice field inside map value struct unsupported error")
	}
}

func userMapSchema() ContractSchema {
	return ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"User": {
					Kind: TypeStruct,
					Name: "User",
					Raw:  "User",
					Fields: []StructField{
						{Name: "Balance", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
						{Name: "Active", Type: TypeRef{Kind: TypeScalar, Raw: "bool", Scalar: "bool"}},
					},
				},
			},
		},
	}
}

func mapStringUserType() TypeRef {
	return mapStringNamedType("User")
}

func mapStringNamedType(name string) TypeRef {
	return TypeRef{
		Kind: TypeMap,
		Raw:  "map[string]" + name,
		Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
		Elem: &TypeRef{Kind: TypeNamed, Raw: name, Name: name},
	}
}

func userStructTypedValue(balance int64, active bool) gno.TypedValue {
	return gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				int64TypedValue(balance),
				boolTypedValue(active),
			},
		},
	}
}

func gnoTypedNilMap() gno.TypedValue {
	return gno.TypedValue{}
}
