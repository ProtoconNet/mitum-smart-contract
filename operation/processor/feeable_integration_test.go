package processor_test

import (
	"context"
	"strings"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	cprocessor "github.com/imfact-labs/currency-model/operation/processor"
	operationtest "github.com/imfact-labs/currency-model/operation/test"
	ccstate "github.com/imfact-labs/currency-model/state/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	modelprocessor "github.com/imfact-labs/smart-contract-model/operation/processor"
)

type feeTestProcessor struct {
	reason base.OperationProcessReasonError
}

func (p feeTestProcessor) PreProcess(
	ctx context.Context,
	_ base.Operation,
	_ base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	return ctx, nil, nil
}

func (p feeTestProcessor) Process(
	_ context.Context,
	_ base.Operation,
	_ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, p.reason, nil
}

func (feeTestProcessor) Close() error { return nil }

func newContractFeeProcessor(
	t *testing.T,
	getStateFunc base.GetStateFunc,
	reason base.OperationProcessReasonError,
) *cprocessor.OperationProcessor {
	t.Helper()

	root := cprocessor.NewOperationProcessor()
	if err := root.SetGetNewProcessorFunc(modelprocessor.GetNewProcessor); err != nil {
		t.Fatalf("set processor resolver: %v", err)
	}

	newProcessor := func(
		base.Height,
		*base.ProposalSignFact,
		base.GetStateFunc,
		base.NewOperationProcessorProcessFunc,
		base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		return feeTestProcessor{reason: reason}, nil
	}
	if err := root.SetProcessorWithProposal(contract.RegisterContractHint, newProcessor); err != nil {
		t.Fatalf("set register processor: %v", err)
	}
	if err := root.SetProcessorWithProposal(contract.CallContractHint, newProcessor); err != nil {
		t.Fatalf("set call processor: %v", err)
	}
	proposalFact := isaac.NewProposalFact(base.GenesisPoint, base.NewStringAddress("feeproposer"), nil, nil)
	var proposal base.ProposalSignFact = isaac.NewProposalSignFact(proposalFact)
	if err := root.SetProposal(&proposal); err != nil {
		t.Fatalf("set proposal: %v", err)
	}

	opr, err := root.New(base.GenesisHeight, getStateFunc, nil, nil)
	if err != nil {
		t.Fatalf("new operation processor: %v", err)
	}

	return opr
}

func setContractFixedFeePolicy(
	tp *operationtest.TestProcessor,
	cid ctypes.CurrencyID,
	receiver base.Address,
	fee int64,
) {
	design := ctypes.NewCurrencyDesign(
		common.ZeroBig,
		cid,
		common.NewBig(9),
		receiver,
		ctypes.NewCurrencyPolicy(common.ZeroBig, ctypes.NewFixedFeeer(receiver, common.NewBig(fee))),
	)
	tp.SetState(common.NewBaseState(
		base.Height(1),
		ccstate.DesignStateKey(cid),
		ccstate.NewCurrencyDesignStateValue(design),
		nil,
		[]util.Hash{},
	), true)
}

func contractFeeTestOperations(
	t *testing.T,
	sender base.Address,
	senderPrivatekey base.Privatekey,
	networkID base.NetworkID,
	cid ctypes.CurrencyID,
) []base.Operation {
	t.Helper()

	register, err := contract.NewRegisterContract(contract.NewRegisterContractFact(
		[]byte("register-fee"),
		sender,
		base.NewStringAddress("registerfeecontract"),
		"package contract",
		map[string]string{"initial": "value"},
		cid,
	))
	if err != nil {
		t.Fatalf("new register operation: %v", err)
	}
	if err := register.Sign(senderPrivatekey, networkID); err != nil {
		t.Fatalf("sign register operation: %v", err)
	}

	call, err := contract.NewCallContract(contract.NewCallContractFactWithItems(
		[]byte("call-fee"),
		sender,
		base.NewStringAddress("callfeecontract"),
		[]contract.CallContractItem{
			contract.NewCallContractItem("Set", map[string]string{"value": "one"}),
			contract.NewCallContractItem("Set", map[string]string{"value": "two"}),
		},
		cid,
	))
	if err != nil {
		t.Fatalf("new call operation: %v", err)
	}
	if err := call.Sign(senderPrivatekey, networkID); err != nil {
		t.Fatalf("sign call operation: %v", err)
	}

	return []base.Operation{register, call}
}

func TestContractOperationsCurrencyFeeIntegration(t *testing.T) {
	for _, operationName := range []string{"register", "call"} {
		t.Run(operationName, func(t *testing.T) {
			getter := operationtest.NewMockStateGetter()
			var tp operationtest.TestProcessor
			tp.Setup(getter)

			sender, _, senderPrivatekey := tp.NewTestAccountState(tp.NewPrivateKey(operationName+"-sender"), true)
			receiver, _, _ := tp.NewTestAccountState(tp.NewPrivateKey(operationName+"-receiver"), true)
			tp.NewTestBalanceState(sender, tp.GenesisCurrency, 100, true)
			tp.NewTestBalanceState(receiver, tp.GenesisCurrency, 0, true)
			setContractFixedFeePolicy(&tp, tp.GenesisCurrency, receiver, 7)

			operations := contractFeeTestOperations(t, sender, senderPrivatekey, tp.NetworkID, tp.GenesisCurrency)
			op := operations[0]
			if operationName == "call" {
				op = operations[1]
			}

			opr := newContractFeeProcessor(t, tp.GetStateFunc, nil)
			merges, reason, err := opr.Process(context.Background(), op, tp.GetStateFunc)
			if err != nil || reason != nil {
				t.Fatalf("process operation: reason=%v err=%v", reason, err)
			}
			if len(merges) != 2 {
				t.Fatalf("fee merge count=%d, want 2", len(merges))
			}

			keys := map[string]bool{}
			for i := range merges {
				keys[merges[i].Key()] = true
			}
			if !keys[ccstate.BalanceStateKey(sender, tp.GenesisCurrency)] {
				t.Fatal("missing fee payer deduction merge")
			}
			if !keys[ccstate.BalanceStateKey(receiver, tp.GenesisCurrency)] {
				t.Fatal("missing fee receiver credit merge")
			}

			receipt, ok := opr.OperationReceipt().(ctypes.CurrencyOperationReceipt)
			if !ok {
				t.Fatalf("receipt type=%T", opr.OperationReceipt())
			}
			fee, ok := receipt.Fee.(ctypes.FixedFeeReceipt)
			if !ok {
				t.Fatalf("fee receipt type=%T", receipt.Fee)
			}
			if fee.Currency() != tp.GenesisCurrency || fee.FeeAmount() != "7" {
				t.Fatalf("unexpected fee receipt: %+v", fee)
			}
		})
	}
}

func TestContractOperationFeeFailurePolicy(t *testing.T) {
	getter := operationtest.NewMockStateGetter()
	var tp operationtest.TestProcessor
	tp.Setup(getter)

	sender, _, senderPrivatekey := tp.NewTestAccountState(tp.NewPrivateKey("failure-sender"), true)
	receiver, _, _ := tp.NewTestAccountState(tp.NewPrivateKey("failure-receiver"), true)
	tp.NewTestBalanceState(sender, tp.GenesisCurrency, 3, true)
	tp.NewTestBalanceState(receiver, tp.GenesisCurrency, 0, true)
	setContractFixedFeePolicy(&tp, tp.GenesisCurrency, receiver, 7)
	op := contractFeeTestOperations(t, sender, senderPrivatekey, tp.NetworkID, tp.GenesisCurrency)[0]

	t.Run("insufficient balance", func(t *testing.T) {
		opr := newContractFeeProcessor(t, tp.GetStateFunc, nil)
		merges, reason, err := opr.Process(context.Background(), op, tp.GetStateFunc)
		if err != nil || reason == nil || len(merges) != 0 {
			t.Fatalf("process operation: merges=%d reason=%v err=%v", len(merges), reason, err)
		}
		if !strings.Contains(reason.Error(), "balance insufficient") {
			t.Fatalf("unexpected insufficient balance reason: %v", reason)
		}
	})

	t.Run("runtime failure does not charge fee", func(t *testing.T) {
		runtimeReason := base.NewBaseOperationProcessReasonError("contract execution failed")
		opr := newContractFeeProcessor(t, tp.GetStateFunc, runtimeReason)
		merges, reason, err := opr.Process(context.Background(), op, tp.GetStateFunc)
		if err != nil || reason == nil || len(merges) != 0 {
			t.Fatalf("process operation: merges=%d reason=%v err=%v", len(merges), reason, err)
		}
		if reason.Error() != runtimeReason.Error() || opr.OperationReceipt() != nil {
			t.Fatalf("runtime failure changed fee/receipt policy: reason=%v receipt=%T", reason, opr.OperationReceipt())
		}
	})

	t.Run("unsupported currency", func(t *testing.T) {
		unsupported := ctypes.CurrencyID("NOPE")
		unsupportedOp := contractFeeTestOperations(t, sender, senderPrivatekey, tp.NetworkID, unsupported)[1]
		opr := newContractFeeProcessor(t, tp.GetStateFunc, nil)
		merges, reason, err := opr.Process(context.Background(), unsupportedOp, tp.GetStateFunc)
		if err != nil || reason == nil || len(merges) != 0 {
			t.Fatalf("process operation: merges=%d reason=%v err=%v", len(merges), reason, err)
		}
		if !strings.Contains(reason.Error(), "currency") {
			t.Fatalf("unexpected unsupported currency reason: %v", reason)
		}
	})
}
