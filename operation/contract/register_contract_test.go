package contract

import (
	"strings"
	"testing"

	types "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

func TestRegisterContractFactAllowsSourceAtMaxSize(t *testing.T) {
	fact := NewRegisterContractFact(
		[]byte("token"),
		base.NewStringAddress("senderlimit0001"),
		base.NewStringAddress("contractlimit0001"),
		strings.Repeat("a", MaxContractSourceBytes),
		map[string]string{},
		types.CurrencyID("ABC"),
	)

	if err := fact.IsValid(nil); err != nil {
		t.Fatalf("IsValid returned error at max size: %v", err)
	}
}

func TestRegisterContractFactRejectsSourceOverMaxSize(t *testing.T) {
	fact := NewRegisterContractFact(
		[]byte("token"),
		base.NewStringAddress("senderlimit0002"),
		base.NewStringAddress("contractlimit0002"),
		strings.Repeat("a", MaxContractSourceBytes+1),
		map[string]string{},
		types.CurrencyID("ABC"),
	)

	err := fact.IsValid(nil)
	if err == nil {
		t.Fatal("expected IsValid to reject oversized contract source")
	}
	if !strings.Contains(err.Error(), "contract source exceeds max size") {
		t.Fatalf("expected max size error, got: %v", err)
	}
}
