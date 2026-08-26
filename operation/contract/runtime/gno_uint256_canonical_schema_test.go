package runtime

import "testing"

func TestUint256CanonicalBoundaryRejectsExternalUintABIAndState(t *testing.T) {
	for _, declaration := range []string{
		"var value u256.Uint\nfunc Initialize(ctx chain.WriteContext) error{return nil}",
		"var value *u256.Uint\nfunc Initialize(ctx chain.WriteContext) error{return nil}",
		"var values []u256.Uint\nfunc Initialize(ctx chain.WriteContext) error{return nil}",
		"var values map[string]*u256.Uint\nfunc Initialize(ctx chain.WriteContext) error{return nil}",
		"type Box struct{ Value *u256.Uint };var box Box\nfunc Initialize(ctx chain.WriteContext) error{return nil}",
		"func Initialize(ctx chain.WriteContext) error{return nil}\nfunc Set(ctx chain.WriteContext,value *u256.Uint) error{return nil}",
		"func Initialize(ctx chain.WriteContext) error{return nil}\nfunc Get(ctx chain.QueryContext) *u256.Uint{return u256.Zero()}",
	} {
		source := "package contract\nimport (\"mitum/math/v1/u256\";\"mitum/chain\")\n" + declaration
		if _, err := AnalyzeContractSchema(source); err == nil {
			t.Fatalf("expected external Uint boundary rejection: %s", declaration)
		}
	}
}

func TestUint256CanonicalBoundaryAllowsLocalCalculationAndStringABI(t *testing.T) {
	source := `package contract
import ("mitum/math/v1/u256";"mitum/chain")
var value string
func Initialize(ctx chain.WriteContext) error { value=u256.ToCanonicalDecimal(u256.Zero());return nil }
func Set(ctx chain.WriteContext,input string) error { parsed,err:=u256.FromCanonicalDecimal(input);if err!=nil{return err};value=u256.ToCanonicalDecimal(parsed);return nil }
func Get(ctx chain.QueryContext) string { return value }`
	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("canonical string boundary rejected: %v", err)
	}
}
