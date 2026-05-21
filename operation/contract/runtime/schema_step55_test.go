package runtime

import "testing"

func TestAnalyzeContractSchemaStructQueryResultSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.WriteContext) error { return nil }
func GetUser(ctx chain.QueryContext, owner string) User { return users[owner] }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaMapQueryResultSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var balances map[string]int64

func Initialize(ctx chain.WriteContext) error { return nil }
func GetBalances(ctx chain.QueryContext) map[string]int64 { return balances }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}
