package contract

import (
	"testing"

	"github.com/imfact-labs/currency-model/operation/extras"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
)

func TestRegisterContractFactFeeAble(t *testing.T) {
	sender := base.NewStringAddress("registerfeesender")
	contract := base.NewStringAddress("registerfeecontract")
	currency := ctypes.CurrencyID("FEE")
	callDataAB := map[string]string{"alpha": "1", "beta": "long-value"}
	callDataBA := map[string]string{"beta": "long-value", "alpha": "1"}

	fact := NewRegisterContractFact([]byte("token"), sender, contract, "package contract", callDataAB, currency)
	reordered := NewRegisterContractFact([]byte("token"), sender, contract, "package contract", callDataBA, currency)
	cid, items, size, hasItems := fact.FeeBase()
	if cid != currency || items != extras.NoItemFeeBaseItemCount || hasItems != extras.HasNoItem {
		t.Fatalf("FeeBase()=(%q,%d,%d,%t)", cid, items, size, hasItems)
	}
	if !fact.FeePayer().Equal(sender) {
		t.Fatalf("FeePayer()=%v, want %v", fact.FeePayer(), sender)
	}
	if _, _, reorderedSize, _ := reordered.FeeBase(); reorderedSize != size {
		t.Fatalf("map insertion order changed fee size: %d != %d", reorderedSize, size)
	}
	if !fact.Hash().Equal(reordered.Hash()) {
		t.Fatal("map insertion order changed existing fact hash")
	}

	longerCode := NewRegisterContractFact([]byte("token"), sender, contract, "package contract\nvar value string", callDataAB, currency)
	if _, _, longerSize, _ := longerCode.FeeBase(); longerSize <= size {
		t.Fatalf("longer source did not increase fee data size: %d <= %d", longerSize, size)
	}
	longerInit := NewRegisterContractFact([]byte("token"), sender, contract, "package contract", map[string]string{"alpha": "a much longer init value"}, currency)
	if _, _, longerSize, _ := longerInit.FeeBase(); longerSize == size {
		t.Fatalf("changed init data did not change fee data size: %d", longerSize)
	}
}

func TestCallContractFactFeeAble(t *testing.T) {
	sender := base.NewStringAddress("callfeesender")
	contract := base.NewStringAddress("callfeecontract")
	currency := ctypes.CurrencyID("FEE")
	items := []CallContractItem{
		NewCallContractItem("First", map[string]string{"alpha": "1", "beta": "two"}),
		NewCallContractItem("Second", map[string]string{"value": "three"}),
	}
	fact := NewCallContractFactWithItems([]byte("token"), sender, contract, items, currency)
	cid, itemCount, size, hasItems := fact.FeeBase()
	if cid != currency || itemCount != len(items) || hasItems != extras.HasItem {
		t.Fatalf("FeeBase()=(%q,%d,%d,%t)", cid, itemCount, size, hasItems)
	}
	if !fact.FeePayer().Equal(sender) {
		t.Fatalf("FeePayer()=%v, want %v", fact.FeePayer(), sender)
	}

	reorderedMap := NewCallContractFactWithItems([]byte("token"), sender, contract, []CallContractItem{
		NewCallContractItem("First", map[string]string{"beta": "two", "alpha": "1"}),
		NewCallContractItem("Second", map[string]string{"value": "three"}),
	}, currency)
	if _, _, reorderedSize, _ := reorderedMap.FeeBase(); reorderedSize != size {
		t.Fatalf("map insertion order changed fee size: %d != %d", reorderedSize, size)
	}
	if !fact.Hash().Equal(reorderedMap.Hash()) {
		t.Fatal("map insertion order changed existing fact hash")
	}

	longerData := NewCallContractFactWithItems([]byte("token"), sender, contract, []CallContractItem{
		NewCallContractItem("First", map[string]string{"alpha": "a much longer call value"}),
		NewCallContractItem("Second", map[string]string{"value": "three"}),
	}, currency)
	if _, _, longerSize, _ := longerData.FeeBase(); longerSize == size {
		t.Fatalf("changed calldata did not change fee data size: %d", longerSize)
	}

	reversed := NewCallContractFactWithItems([]byte("token"), sender, contract, []CallContractItem{items[1], items[0]}, currency)
	if fact.Hash().Equal(reversed.Hash()) {
		t.Fatal("call item order must remain part of fact semantics")
	}
}
