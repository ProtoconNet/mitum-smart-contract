package runtime

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestTopLevelScalarSliceRoundTripPreservesOrder(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeSlice,
		Raw:  "[]string",
		Elem: &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
	}

	value, err := ExtractSnapshotValue(schema, typ, sliceTypedValue(
		stringTypedValue("alice"),
		stringTypedValue("bob"),
		stringTypedValue("carol"),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	if value.Kind != string(TypeSlice) || len(value.Items) != 3 {
		t.Fatalf("unexpected slice snapshot value: %#v", value)
	}
	if value.Items[0].Scalar != "alice" || value.Items[1].Scalar != "bob" || value.Items[2].Scalar != "carol" {
		t.Fatalf("unexpected slice order: %#v", value.Items)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `[]string{"alice","bob","carol"}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestTopLevelNamedStructSliceRoundTripPreservesOrder(t *testing.T) {
	schema := sliceStructSchema()
	typ := TypeRef{
		Kind: TypeSlice,
		Raw:  "[]User",
		Elem: &TypeRef{Kind: TypeNamed, Raw: "User", Name: "User"},
	}

	valueA, err := ExtractSnapshotValue(schema, typ, sliceTypedValue(
		userStructTypedValue(10, true),
		userStructTypedValue(20, false),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}
	valueB, err := ExtractSnapshotValue(schema, typ, sliceTypedValue(
		userStructTypedValue(10, true),
		userStructTypedValue(20, false),
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
		t.Fatalf("expected deterministic bytes for same slice state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	lit, err := BuildLiteral(schema, typ, valueA)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `[]User{User{Balance:10,Active:true},User{Balance:20,Active:false}}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestStructFieldSliceRoundTrip(t *testing.T) {
	schema := configWithSlicesSchema()
	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}

	value, err := ExtractSnapshotValue(schema, typ, configWithSlicesTypedValue(
		sliceTypedValue(
			stringTypedValue("alice"),
			stringTypedValue("bob"),
		),
		sliceTypedValue(
			nestedUserStructTypedValue(10, true, 100),
			nestedUserStructTypedValue(20, false, 200),
		),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Names:[]string{"alice","bob"},Users:[]User{User{Balance:10,Meta:Meta{Active:true,Limit:100}},User{Balance:20,Meta:Meta{Active:false,Limit:200}}}}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestSliceNilAndEmptyPolicy(t *testing.T) {
	schema := configWithSlicesSchema()
	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}

	value, err := ExtractSnapshotValue(schema, typ, configWithSlicesTypedValue(
		gno.TypedValue{},
		emptySliceTypedValue(),
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	if len(value.Fields) != 2 {
		t.Fatalf("unexpected field count: %#v", value.Fields)
	}
	if !value.Fields[0].Value.IsNil {
		t.Fatalf("expected nil slice field for Names, got %#v", value.Fields[0].Value)
	}
	if value.Fields[1].Value.IsNil {
		t.Fatalf("expected empty non-nil slice field for Users, got %#v", value.Fields[1].Value)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Names:nil,Users:[]User{}}` {
		t.Fatalf("unexpected nil/empty literal: %s", lit)
	}
}

func TestSliceNestedSliceUnsupported(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeSlice,
		Raw:  "[][]string",
		Elem: &TypeRef{
			Kind: TypeSlice,
			Raw:  "[]string",
			Elem: &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
		},
	}

	_, err := ExtractSnapshotValue(schema, typ, sliceTypedValue())
	if err == nil {
		t.Fatalf("expected [][]string unsupported error")
	}
}

func TestSliceMapElemUnsupported(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeSlice,
		Raw:  "[]map[string]int64",
		Elem: &TypeRef{
			Kind: TypeMap,
			Raw:  "map[string]int64",
			Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
			Elem: &TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"},
		},
	}

	_, err := ExtractSnapshotValue(schema, typ, sliceTypedValue())
	if err == nil {
		t.Fatalf("expected []map[string]T unsupported error")
	}
}

func TestSliceAnonymousStructElemUnsupported(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeSlice,
		Raw:  "[]struct{Count int64}",
		Elem: &TypeRef{
			Kind: TypeStruct,
			Raw:  "struct{Count int64}",
			Fields: []StructField{
				{Name: "Count", Type: TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, typ, sliceTypedValue())
	if err == nil {
		t.Fatalf("expected []anonymous-struct unsupported error")
	}
}

func sliceStructSchema() ContractSchema {
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

func configWithSlicesSchema() ContractSchema {
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
						{Name: "Names", Type: TypeRef{
							Kind: TypeSlice,
							Raw:  "[]string",
							Elem: &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
						}},
						{Name: "Users", Type: TypeRef{
							Kind: TypeSlice,
							Raw:  "[]User",
							Elem: &TypeRef{Kind: TypeNamed, Raw: "User", Name: "User"},
						}},
					},
				},
			},
		},
	}
}

func configWithSlicesTypedValue(names gno.TypedValue, users gno.TypedValue) gno.TypedValue {
	return gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				names,
				users,
			},
		},
	}
}

func sliceTypedValue(items ...gno.TypedValue) gno.TypedValue {
	alloc := gno.NewAllocator(1024)
	list := append([]gno.TypedValue(nil), items...)
	return gno.TypedValue{V: alloc.NewSliceFromList(list)}
}

func emptySliceTypedValue() gno.TypedValue {
	alloc := gno.NewAllocator(1024)
	return gno.TypedValue{V: alloc.NewSliceFromList([]gno.TypedValue{})}
}
