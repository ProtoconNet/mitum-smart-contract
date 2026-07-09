package processor

import (
	"testing"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/currency-model/operation/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
)

func TestExecutionClassRegisterContractSerial(t *testing.T) {
	fact := contract.NewRegisterContractFact(
		[]byte("token"),
		base.NewStringAddress("classsender0001"),
		base.NewStringAddress("classcontract001"),
		"package main",
		map[string]string{},
		ctypes.CurrencyID("ABC"),
	)
	op, err := contract.NewRegisterContract(fact)
	if err != nil {
		t.Fatalf("NewRegisterContract returned error: %v", err)
	}

	if got := ExecutionClass(op); got != isaac.OperationExecutionClassSerial {
		t.Fatalf("unexpected execution class: got %q, want %q", got, isaac.OperationExecutionClassSerial)
	}
}

func TestExecutionClassCallContractSerial(t *testing.T) {
	fact := contract.NewCallContractFact(
		[]byte("token"),
		base.NewStringAddress("classsender0002"),
		base.NewStringAddress("classcontract002"),
		map[string]string{"function": "Update"},
		ctypes.CurrencyID("ABC"),
	)
	op, err := contract.NewCallContract(fact)
	if err != nil {
		t.Fatalf("NewCallContract returned error: %v", err)
	}

	if got := ExecutionClass(op); got != isaac.OperationExecutionClassSerial {
		t.Fatalf("unexpected execution class: got %q, want %q", got, isaac.OperationExecutionClassSerial)
	}
}

func TestExecutionClassCurrencyOperationParallel(t *testing.T) {
	fact := currency.NewTransferFact(
		[]byte("token"),
		base.NewStringAddress("classsender0003"),
		[]currency.TransferItem{
			currency.NewTransferItemSingleAmount(
				base.NewStringAddress("classreceiver001"),
				ctypes.NewAmount(common.NewBig(1), ctypes.CurrencyID("ABC")),
			),
		},
		ctypes.CurrencyID("ABC"),
	)
	op, err := currency.NewTransfer(fact)
	if err != nil {
		t.Fatalf("NewTransfer returned error: %v", err)
	}

	if got := ExecutionClass(op); got != isaac.OperationExecutionClassParallel {
		t.Fatalf("unexpected execution class: got %q, want %q", got, isaac.OperationExecutionClassParallel)
	}
}

func TestExecutionClassUnknownOperationParallel(t *testing.T) {
	fact := base.NewBaseFact(hint.MustNewHint("unknown-operation-fact-v0.0.1"), []byte("token"))
	op := base.NewBaseOperation(hint.MustNewHint("unknown-operation-v0.0.1"), fact)

	if got := ExecutionClass(op); got != isaac.OperationExecutionClassParallel {
		t.Fatalf("unexpected execution class: got %q, want %q", got, isaac.OperationExecutionClassParallel)
	}
}
