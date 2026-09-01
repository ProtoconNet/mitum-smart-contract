package spec

import (
	"testing"

	"github.com/imfact-labs/smart-contract-model/operation/contract"
)

func TestSmartContractHinterInventory(t *testing.T) {
	wantHinters := map[string]bool{
		contract.RegisterContractHint.String(): false,
		contract.CallContractHint.String():     false,
		contract.CallContractItemHint.String(): false,
	}
	for i := range AddedHinters {
		if _, found := wantHinters[AddedHinters[i].Hint.String()]; found {
			wantHinters[AddedHinters[i].Hint.String()] = true
		}
	}
	for ht, found := range wantHinters {
		if !found {
			t.Fatalf("missing smart-contract hinter %q", ht)
		}
	}

	wantFacts := map[string]bool{
		contract.RegisterContractFactHint.String(): false,
		contract.CallContractFactHint.String():     false,
	}
	if len(AddedSupportedHinters) != len(wantFacts) {
		t.Fatalf("supported fact count=%d, want %d", len(AddedSupportedHinters), len(wantFacts))
	}
	for i := range AddedSupportedHinters {
		wantFacts[AddedSupportedHinters[i].Hint.String()] = true
	}
	for ht, found := range wantFacts {
		if !found {
			t.Fatalf("missing supported fact hinter %q", ht)
		}
	}
}
