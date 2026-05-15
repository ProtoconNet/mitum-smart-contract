package runtime

import (
	"encoding/json"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const flatStructContractSource = `package contract
import "mitum/chain"

type Config struct {
	Owner string
	Paused bool
	Limit int64
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
`

func TestExtractSnapshotValueFlatStruct(t *testing.T) {
	schema, err := AnalyzeContractSchema(flatStructContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	value, err := ExtractSnapshotValue(schema, TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}, gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				stringTypedValue("alice"),
				boolTypedValue(true),
				int64TypedValue(10),
			},
		},
	})
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}

	if value.Kind != string(TypeStruct) {
		t.Fatalf("expected struct snapshot value, got %q", value.Kind)
	}
	if len(value.Fields) != 3 {
		t.Fatalf("expected 3 snapshot fields, got %d", len(value.Fields))
	}
	if value.Fields[0].Name != "Owner" || value.Fields[0].Value.Scalar != "alice" {
		t.Fatalf("unexpected first field: %#v", value.Fields[0])
	}
	if value.Fields[1].Name != "Paused" || value.Fields[1].Value.Scalar != "true" {
		t.Fatalf("unexpected second field: %#v", value.Fields[1])
	}
	if value.Fields[2].Name != "Limit" || value.Fields[2].Value.Scalar != "10" {
		t.Fatalf("unexpected third field: %#v", value.Fields[2])
	}
}

func TestFlatStructSnapshotRoundTrip(t *testing.T) {
	schema, err := AnalyzeContractSchema(flatStructContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	typ := TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}
	tv := gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				stringTypedValue("alice"),
				boolTypedValue(false),
				int64TypedValue(10),
			},
		},
	}

	valueA, err := ExtractSnapshotValue(schema, typ, tv)
	if err != nil {
		t.Fatalf("ExtractSnapshotValue returned error: %v", err)
	}
	valueB, err := ExtractSnapshotValue(schema, typ, tv)
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
		t.Fatalf("expected deterministic bytes for same struct state:\nfirst:  %s\nsecond: %s", bytesA, bytesB)
	}

	lit, err := BuildLiteral(schema, typ, valueA)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}
	if lit != `Config{Owner:"alice",Paused:false,Limit:10}` {
		t.Fatalf("unexpected round-trip literal: %s", lit)
	}
}

func TestExtractSnapshotValueNestedStructUnsupported(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Config": {
					Kind: TypeStruct,
					Name: "Config",
					Raw:  "Config",
					Fields: []StructField{
						{Name: "Limits", Type: TypeRef{Kind: TypeNamed, Name: "Limits", Raw: "Limits"}},
					},
				},
				"Limits": {
					Kind: TypeStruct,
					Name: "Limits",
					Raw:  "Limits",
					Fields: []StructField{
						{Name: "Daily", Type: TypeRef{Kind: TypeScalar, Scalar: "int64", Raw: "int64"}},
					},
				},
			},
		},
	}

	_, err := ExtractSnapshotValue(schema, TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}, gno.TypedValue{
		V: &gno.StructValue{
			Fields: []gno.TypedValue{
				{V: &gno.StructValue{Fields: []gno.TypedValue{int64TypedValue(1)}}},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected nested struct unsupported error")
	}

	if got := err.Error(); !strings.Contains(got, "flat struct globals require scalar fields only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func stringTypedValue(v string) gno.TypedValue {
	return gno.TypedValue{V: gno.StringValue(v)}
}

func boolTypedValue(v bool) gno.TypedValue {
	var tv gno.TypedValue
	tv.SetBool(v)
	return tv
}

func int64TypedValue(v int64) gno.TypedValue {
	var tv gno.TypedValue
	tv.SetInt64(v)
	return tv
}

func mustMarshalSnapshotDoc(t *testing.T, doc SnapshotDoc) []byte {
	t.Helper()

	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	return b
}
