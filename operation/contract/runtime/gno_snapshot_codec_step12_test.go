package runtime

import (
	"testing"
)

func TestBuildLiteralScalarStillWorks(t *testing.T) {
	schema := ContractSchema{}
	typ := TypeRef{Kind: TypeScalar, Scalar: "string", Raw: "string"}
	value := SnapshotValue{
		Kind:   string(TypeScalar),
		Scalar: "hello",
	}

	lit, err := BuildLiteral(schema, typ, value)
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}

	if lit != `"hello"` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}

func TestBuildLiteralFlatStructWorks(t *testing.T) {
	schema := ContractSchema{
		Types: TypeRegistry{
			Structs: map[string]TypeRef{
				"Config": {
					Kind: TypeStruct,
					Name: "Config",
					Raw:  "Config",
					Fields: []StructField{
						{Name: "Owner", Type: TypeRef{Kind: TypeScalar, Scalar: "string", Raw: "string"}},
					},
				},
			},
		},
	}

	lit, err := BuildLiteral(schema, TypeRef{Kind: TypeNamed, Name: "Config", Raw: "Config"}, SnapshotValue{
		Kind: string(TypeStruct),
		Fields: []SnapshotField{
			{
				Name: "Owner",
				Value: SnapshotValue{
					Kind:   string(TypeScalar),
					Scalar: "alice",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildLiteral returned error: %v", err)
	}

	if lit != `Config{Owner:"alice"}` {
		t.Fatalf("unexpected literal: %s", lit)
	}
}
