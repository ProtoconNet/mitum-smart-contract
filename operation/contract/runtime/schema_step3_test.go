package runtime

import "testing"

func TestAnalyzeContractSchemaStructFieldMapStringScalarSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Users map[string]int64
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaStructFieldSliceScalarSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Flags []string
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}
