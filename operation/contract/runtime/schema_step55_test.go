package runtime

import "testing"

func TestAnalyzeContractSchemaStructQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func GetUser(ctx chain.ContractContext, owner string) User { return users[owner] }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected struct query result unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "GetUser", "query", "scalar") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var balances map[string]int64

func Initialize(ctx chain.ContractContext) error { return nil }
func GetBalances(ctx chain.ContractContext) map[string]int64 { return balances }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map query result unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "GetBalances", "query", "scalar") {
		t.Fatalf("unexpected error: %v", err)
	}
}
