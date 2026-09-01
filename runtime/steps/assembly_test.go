package steps

import (
	"testing"

	"github.com/imfact-labs/currency-model/operation/currency"
	jsonenc "github.com/imfact-labs/mitum2/util/encoder/json"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	"github.com/imfact-labs/smart-contract-model/runtime/spec"
)

func TestProposalFactFilterComposesCurrencyAndSmartContractFacts(t *testing.T) {
	f := IsSupportedProposalOperationFactHintFunc()
	for _, ht := range []struct {
		name string
		ok   bool
	}{
		{contract.RegisterContractFactHint.String(), f(contract.RegisterContractFactHint)},
		{contract.CallContractFactHint.String(), f(contract.CallContractFactHint)},
		{currency.TransferFactHint.String(), f(currency.TransferFactHint)},
	} {
		if !ht.ok {
			t.Fatalf("supported fact rejected: %s", ht.name)
		}
	}
	if f(contract.CallContractItemHint) {
		t.Fatal("call item hint must not be a proposal fact hint")
	}
}

func TestSmartContractHintersDecodeOperationsAndCallItem(t *testing.T) {
	enc := jsonenc.NewEncoder()
	for i := range spec.AddedHinters {
		if err := enc.Add(spec.AddedHinters[i]); err != nil {
			t.Fatal(err)
		}
	}
	for i := range spec.AddedSupportedHinters {
		if err := enc.Add(spec.AddedSupportedHinters[i]); err != nil {
			t.Fatal(err)
		}
	}

	item := contract.NewCallContractItem("Update", map[string]string{"value": "next"})
	b, err := item.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := enc.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded.(contract.CallContractItem); !ok {
		t.Fatalf("decoded call item type=%T", decoded)
	}
}

func TestOperationProcessorHintsAreRegisterAndCall(t *testing.T) {
	hints := OperationProcessorHints()
	if len(hints) != 2 {
		t.Fatalf("processor hint count=%d", len(hints))
	}
	if !hints[0].Equal(contract.RegisterContractHint) || !hints[1].Equal(contract.CallContractHint) {
		t.Fatalf("processor hints=%v", hints)
	}
}
