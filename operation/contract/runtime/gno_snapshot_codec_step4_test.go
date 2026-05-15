package runtime

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestExtractSnapshotValueMapStringScalarDeterministicOrder(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeMap,
		Raw:  "map[string]int64",
		Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
		Elem: &TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"},
	}

	valueA, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: int64TypedValueWithType(20)},
		mapTestEntry{key: "alice", value: int64TypedValueWithType(10)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	valueB, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "alice", value: int64TypedValueWithType(10)},
		mapTestEntry{key: "bob", value: int64TypedValueWithType(20)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	bytesA := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "balances", Value: valueA}},
	})
	bytesB := mustMarshalSnapshotDoc(t, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "balances", Value: valueB}},
	})

	if string(bytesA) != string(bytesB) {
		t.Fatalf("expected deterministic bytes for same logical map state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	if len(valueA.Entries) != 2 || valueA.Entries[0].Key != "alice" || valueA.Entries[1].Key != "bob" {
		t.Fatalf("expected lexical key order, got %#v", valueA.Entries)
	}
}

func TestBuildLiteralMapStringScalarRoundTrip(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeMap,
		Raw:  "map[string]int64",
		Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
		Elem: &TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"},
	}

	value, err := ExtractSnapshotValue(schema, typ, mapTypedValue(
		mapTestEntry{key: "bob", value: int64TypedValueWithType(20)},
		mapTestEntry{key: "alice", value: int64TypedValueWithType(10)},
	))
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}

	if lit != `map[string]int64{"alice":10,"bob":20}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestExtractSnapshotValueNilAndEmptyMapPolicy(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{
		Kind: TypeMap,
		Raw:  "map[string]bool",
		Key:  &TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"},
		Elem: &TypeRef{Kind: TypeScalar, Raw: "bool", Scalar: "bool"},
	}

	nilValue, err := ExtractSnapshotValue(schema, typ, gno.TypedValue{})
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
	if emptyLit != `map[string]bool{}` {
		t.Fatalf("unexpected empty map literal: %s", emptyLit)
	}
}

type mapTestEntry struct {
	key   string
	value gno.TypedValue
}

func mapTypedValue(entries ...mapTestEntry) gno.TypedValue {
	alloc := gno.NewAllocator(1024)
	mv := &gno.MapValue{}
	mv.MakeMap(len(entries))

	for _, entry := range entries {
		key := gno.TypedValue{T: gno.StringType, V: gno.StringValue(entry.key)}
		ptr := mv.GetPointerForKey(alloc, nil, key)
		*ptr.TV = entry.value
	}

	return gno.TypedValue{V: mv}
}

func emptyMapTypedValue() gno.TypedValue {
	mv := &gno.MapValue{}
	mv.MakeMap(0)
	return gno.TypedValue{V: mv}
}

func int64TypedValueWithType(v int64) gno.TypedValue {
	tv := int64TypedValue(v)
	tv.T = gno.Int64Type
	return tv
}
