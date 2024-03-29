package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestUpdateCurrencyProcessor struct {
	test.TestProcessor
	receiver *base.Address
	amount   *types.Amount
	currency *types.CurrencyID
	policy   *types.CurrencyPolicy
	op       UpdateCurrency
}

func (t *TestUpdateCurrencyProcessor) Receiver() base.Address {
	return *t.receiver
}

func (t *TestUpdateCurrencyProcessor) Amount() types.Amount {
	return *t.amount
}

func (t *TestUpdateCurrencyProcessor) Currency() types.CurrencyID {
	return *t.currency
}

func (t *TestUpdateCurrencyProcessor) Policy() types.CurrencyPolicy {
	return *t.policy
}

func (t *TestUpdateCurrencyProcessor) Create() *TestUpdateCurrencyProcessor {
	t.Opr, _ = NewUpdateCurrencyProcessor(base.MaxThreshold)(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)
	return t
}

func (t *TestUpdateCurrencyProcessor) SetCurrency(cid string, am int64, addr base.Address, instate bool) *TestUpdateCurrencyProcessor {
	t.NewTestCurrencyState(cid, addr, instate)
	t.NewTestBalanceState(addr, types.CurrencyID(cid), am, instate)
	c := types.CurrencyID(cid)
	t.currency = &c
	return t
}

func (t *TestUpdateCurrencyProcessor) SetCurrencyPolicy(am int64) *TestUpdateCurrencyProcessor {
	p := types.NewCurrencyPolicy(common.NewBig(am), types.NewNilFeeer())
	t.policy = &p
	return t
}

func (t *TestUpdateCurrencyProcessor) SetReceiver(receiverPriv string, inState bool) *TestUpdateCurrencyProcessor {
	receiverAddr, _, _ := t.NewTestAccountState(receiverPriv, inState)
	t.receiver = &receiverAddr

	return t
}

func (t *TestUpdateCurrencyProcessor) SetAmount(am int64, cid types.CurrencyID) *TestUpdateCurrencyProcessor {
	a := types.NewAmount(common.NewBig(am), cid)
	t.amount = &a
	return t
}

func (t *TestUpdateCurrencyProcessor) MakeOperation() *TestUpdateCurrencyProcessor {
	if t.currency == nil {
		panic("execute SetCurrency")
	}
	if t.policy == nil {
		panic("execute SetCurrencyPolicy")
	}

	op, _ := NewUpdateCurrency(NewUpdateCurrencyFact([]byte("token"), t.Currency(), t.Policy()))
	_ = op.NodeSign(t.NodePriv, t.NetworkID, t.NodeAddr)
	t.op = op

	return t
}

func (t *TestUpdateCurrencyProcessor) LoadOperation(fileName string) *TestUpdateCurrencyProcessor {
	var ok bool
	op := t.TestProcessor.LoadOperation(fileName)
	t.op, ok = op.(UpdateCurrency)
	if !ok {
		panic("operation type is not UpdateCurrency")
	}

	return t
}

func (t *TestUpdateCurrencyProcessor) Print(fileName string) *TestUpdateCurrencyProcessor {
	t.TestProcessor.Print(fileName, t.op)

	return t
}

func (t *TestUpdateCurrencyProcessor) RunPreProcess() error {
	//t.MockGetter.On("Get", mock.Anything).Return(nil, false, nil)
	_, err, _ := t.Opr.PreProcess(context.Background(), t.op, t.GetStateFunc)

	return err
}

func (t *TestUpdateCurrencyProcessor) IsValid() error {
	err := t.op.IsValid(t.NetworkID)

	return err
}

//func (t *TestUpdateCurrencyProcessor) RunPreProcess() *TestUpdateCurrencyProcessor {
//	t.BaseTestOperationProcessorNoItem.RunPreProcess()
//
//	return t
//}
//
//func (t *TestUpdateCurrencyProcessor) RunProcess() *TestUpdateCurrencyProcessor {
//	t.BaseTestOperationProcessorNoItem.RunProcess()
//
//	return t
//}
