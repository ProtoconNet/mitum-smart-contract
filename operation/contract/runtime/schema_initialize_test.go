package runtime

import "testing"

func TestAnalyzeContractSchemaInitializeScalarArgsSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext, owner string, label string, limit int64) error { return nil }
func GetLabel(ctx chain.ContractContext) string { return "" }
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}

func TestAnalyzeContractSchemaInitializeCompositeArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct { Owner string }

func Initialize(ctx chain.ContractContext, cfg Config) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected composite initialize arg rejection")
	}
	if got := err.Error(); got == "" || !containsAll(got, "invalid initialize parameter type", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaInitializeNoArgsStillSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}
