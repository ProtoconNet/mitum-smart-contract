package runtime

import "testing"

func TestAnalyzeContractSchemaScalarWriteArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var label string

func Initialize(ctx chain.WriteContext) error { return nil }
func SetLabel(ctx chain.WriteContext, next string) error { return nil }
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}

func TestAnalyzeContractSchemaStructWriteArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct { Owner string }

func Initialize(ctx chain.WriteContext) error { return nil }
func SetConfig(ctx chain.WriteContext, cfg Config) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected struct write arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetConfig", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapWriteArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func SetFlags(ctx chain.WriteContext, flags map[string]bool) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map write arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetFlags", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceWriteArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func SetAliases(ctx chain.WriteContext, aliases []string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected slice write arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetAliases", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}
