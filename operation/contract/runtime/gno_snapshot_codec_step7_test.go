package runtime

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestStructFieldMapRoundTripDeterministicBytes(t *testing.T) {
	schema := configWithMapsSchema()
	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}

	valueA, err := ExtractSnapshotValue(schema, typ, configWithMapsTypedValue(
		mapTypedValue(
			mapTestEntry{key: "beta", value: boolTypedValue(false)},
			mapTestEntry{key: "alpha", value: boolTypedValue(true)},
		),
		mapTypedValue(
			mapTestEntry{key: "bob", value: userWithMetaTypedValue(20, false, 200)},
			mapTestEntry{key: "alice", value: userWithMetaTypedValue(10, true, 100)},
		),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	valueB, err := ExtractSnapshotValue(schema, typ, configWithMapsTypedValue(
		mapTypedValue(
			mapTestEntry{key: "alpha", value: boolTypedValue(true)},
			mapTestEntry{key: "beta", value: boolTypedValue(false)},
		),
		mapTypedValue(
			mapTestEntry{key: "alice", value: userWithMetaTypedValue(10, true, 100)},
			mapTestEntry{key: "bob", value: userWithMetaTypedValue(20, false, 200)},
		),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	bytesA := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "config", Value: valueA}},
	})
	bytesB := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "config", Value: valueB}},
	})

	if string(bytesA) != string(bytesB) {
		t.Fatalf("expected deterministic bytes for same logical struct-with-map state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	lit, err := BuildLiteral(schema, typ, valueA)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Flags:map[string]bool{"alpha":true,"beta":false},Users:map[string]User{"alice":User{Balance:10,Meta:Meta{Active:true,Limit:100}},"bob":User{Balance:20,Meta:Meta{Active:false,Limit:200}}}}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestStructFieldMapNilAndEmptyMapPolicy(t *testing.T) {
	schema := configWithMapsSchema()
	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}

	value, err := ExtractSnapshotValue(schema, typ, configWithMapsTypedValue(
		gno.TypedValue{},
		emptyMapTypedValue(),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	if len(value.Fields) != 2 {
		t.Fatalf("unexpected field count: %#v", value.Fields)
	}
	if !value.Fields[0].Value.IsNil {
		t.Fatalf("expected nil map field for Flags, got %#v", value.Fields[0].Value)
	}
	if value.Fields[1].Value.IsNil {
		t.Fatalf("expected empty non-nil map field for Users, got %#v", value.Fields[1].Value)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Flags:nil,Users:map[string]User{}}` {
		t.Fatalf("unexpected nil/empty literal: %s", lit)
	}
}

func TestStructFieldMapStringSliceElemUnsupported(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Config": {
					Kind: TypeStruct,
					Name: "Config",
					Raw:  "Config",
					Fields: []StructField{
						{Name: "Extra", Type: TypeRef{
							Kind: TypeMap,
							Raw:  "map[string][]string",
							Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
							Elem: &TypeRef{Kind: TypeSlice, Raw: "[]string", Elem: &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"}},
						}},
					},
				},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}, gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{emptyMapTypedValue()},
		},
	})
	if err == nil {
		t.Fatalf("expected map[string][]T unsupported error")
	}
}

func configWithMapsSchema() ContractSchema {
	return ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Meta": {
					Kind: TypeStruct,
					Name: "Meta",
					Raw:  "Meta",
					Fields: []StructField{
						{Name: "Active", Type: TypeRef{Kind: TypeScalar, Raw: "bool", Scalar: "bool"}},
						{Name: "Limit", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
					},
				},
				"User": {
					Kind: TypeStruct,
					Name: "User",
					Raw:  "User",
					Fields: []StructField{
						{Name: "Balance", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
						{Name: "Meta", Type: TypeRef{Kind: TypeNamed, Raw: "Meta", Name: "Meta"}},
					},
				},
				"Config": {
					Kind: TypeStruct,
					Name: "Config",
					Raw:  "Config",
					Fields: []StructField{
						{Name: "Flags", Type: TypeRef{
							Kind: TypeMap,
							Raw:  "map[string]bool",
							Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
							Elem: &TypeRef{Kind: TypeScalar, Raw: "bool", Scalar: "bool"},
						}},
						{Name: "Users", Type: mapStringNamedType("User")},
					},
				},
			},
		},
	}
}

func configWithMapsTypedValue(flags gno.TypedValue, users gno.TypedValue) gno.TypedValue {
	return gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				flags,
				users,
			},
		},
	}
}

func userWithMetaTypedValue(balance int64, active bool, limit int64) gno.TypedValue {
	return gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				int64TypedValue(balance),
				{
					V: &gno.StructValue{
						Fields: []gno.TypedValue{
							boolTypedValue(active),
							int64TypedValue(limit),
						},
					},
				},
			},
		},
	}
}
