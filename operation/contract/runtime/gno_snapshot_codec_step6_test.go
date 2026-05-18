package runtime

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestExtractSnapshotValueNestedStruct(t *testing.T) {
	schema := nestedConfigSchema()

	value, err := ExtractSnapshotValue(schema, TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}, nestedConfigTypedValue("alice", 10, 20))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	if value.Kind != string(TypeStruct) || len(value.Fields) != 2 {
		t.Fatalf("unexpected nested struct snapshot value: %#v", value)
	}
	if value.Fields[0].Name != "Owner" || value.Fields[0].Value.Scalar != "alice" {
		t.Fatalf("unexpected Owner field: %#v", value.Fields[0])
	}
	if value.Fields[1].Name != "Limits" || value.Fields[1].Value.Kind != string(TypeStruct) {
		t.Fatalf("unexpected Limits field: %#v", value.Fields[1])
	}
	if len(value.Fields[1].Value.Fields) != 2 {
		t.Fatalf("unexpected nested Limits fields: %#v", value.Fields[1].Value.Fields)
	}
}

func TestBuildLiteralNestedStructRoundTrip(t *testing.T) {
	schema := nestedConfigSchema()
	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}

	valueA, err := ExtractSnapshotValue(schema, typ, nestedConfigTypedValue("alice", 10, 20))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}
	valueB, err := ExtractSnapshotValue(schema, typ, nestedConfigTypedValue("alice", 10, 20))
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
		t.Fatalf("expected deterministic bytes for same nested struct state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	lit, err := BuildLiteral(schema, typ, valueA)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Owner:"alice",Limits:Limits{Daily:10,Max:20}}` {
		t.Fatalf("unexpected nested struct literal: %s", lit)
	}
}

func TestExtractSnapshotValueMapStringNestedStructDeterministicOrder(t *testing.T) {
	schema := nestedUserMapSchema()
	typ := mapStringNamedType("User")

	valueA, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: nestedUserStructTypedValue(20, false, 200)},
		mapTestEntry{key: "alice", value: nestedUserStructTypedValue(10, true, 100)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}
	valueB, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "alice", value: nestedUserStructTypedValue(10, true, 100)},
		mapTestEntry{key: "bob", value: nestedUserStructTypedValue(20, false, 200)},
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
		t.Fatalf("expected deterministic bytes for same logical nested map state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}
}

func TestBuildLiteralMapStringNestedStructRoundTrip(t *testing.T) {
	schema := nestedUserMapSchema()
	typ := mapStringNamedType("User")

	value, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: nestedUserStructTypedValue(20, false, 200)},
		mapTestEntry{key: "alice", value: nestedUserStructTypedValue(10, true, 100)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}

	if lit != `map[string]User{"alice":User{Balance:10,Meta:Meta{Active:true,Limit:100}},"bob":User{Balance:20,Meta:Meta{Active:false,Limit:200}}}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestExtractSnapshotValueMapStringNestedStructNilAndEmptyMapPolicy(t *testing.T) {
	schema := nestedUserMapSchema()
	typ := mapStringNamedType("User")

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

func nestedConfigSchema() ContractSchema {
	return ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Limits": {
					Kind: TypeStruct,
					Name: "Limits",
					Raw:  "Limits",
					Fields: []StructField{
						{Name: "Daily", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
						{Name: "Max", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
					},
				},
				"Config": {
					Kind: TypeStruct,
					Name: "Config",
					Raw:  "Config",
					Fields: []StructField{
						{Name: "Owner", Type: TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"}},
						{Name: "Limits", Type: TypeRef{Kind: TypeNamed, Raw: "Limits", Name: "Limits"}},
					},
				},
			},
		},
	}
}

func nestedUserMapSchema() ContractSchema {
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
			},
		},
	}
}

func nestedConfigTypedValue(owner string, daily int64, max int64) gno.TypedValue {
	return gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				stringTypedValue(owner),
				{
					V: &gno.StructValue{
						Fields: []gno.TypedValue{
							int64TypedValue(daily),
							int64TypedValue(max),
						},
					},
				},
			},
		},
	}
}

func nestedUserStructTypedValue(balance int64, active bool, limit int64) gno.TypedValue {
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
