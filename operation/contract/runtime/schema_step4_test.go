package runtime

import "testing"

func TestAnalyzeContractSchemaMapStringScalarGlobalSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var balances map[string]int64
var flags map[string]bool
var names map[string]string

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, owner string, amount int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 3 {
		t.Fatalf("expected 3 persistent globals, got %d", len(schema.PersistentGlobals))
	}

	for _, binding := range schema.PersistentGlobals {
		if binding.Type.Kind != TypeMap {
			t.Fatalf("expected map global type, got %#v", binding.Type)
		}
	}
}

func TestAnalyzeContractSchemaNonStringMapKeyUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var balances map[int64]string

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, owner string, amount int64) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected non-string map key unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "balances", "int64", MapGlobalSupportDescription) {
		t.Fatalf("unexpected error: %v", err)
	}
}
