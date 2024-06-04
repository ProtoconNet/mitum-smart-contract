package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestWithdrawProcessor struct {
	*test.BaseTestOperationProcessorWithItem[Withdraw, WithdrawItem]
}

func NewTestWithdrawProcessor(
	tp *test.TestProcessor,
) TestWithdrawProcessor {
	t := test.NewBaseTestOperationProcessorWithItem[Withdraw, WithdrawItem](tp)
	return TestWithdrawProcessor{&t}
}

func (t *TestWithdrawProcessor) Create() *TestWithdrawProcessor {
	t.Opr, _ = NewWithdrawProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)
	return t
}

func (t *TestWithdrawProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestWithdrawProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.SetAmount(am, cid, target)

	return t
}

func (t *TestWithdrawProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestWithdrawProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestWithdrawProcessor) LoadOperation(fileName string) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.LoadOperation(fileName)

	return t
}

func (t *TestWithdrawProcessor) Print(fileName string) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.Print(fileName)

	return t
}

func (t *TestWithdrawProcessor) MakeItem(
	target test.Account, amounts []types.Amount, targetItems []WithdrawItem,
) *TestWithdrawProcessor {
	item := NewWithdrawItemMultiAmounts(target.Address(), amounts)
	test.UpdateSlice[WithdrawItem](item, targetItems)

	return t
}

func (t *TestWithdrawProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, items []WithdrawItem,
) *TestWithdrawProcessor {
	op, _ := NewWithdraw(NewWithdrawFact([]byte("token"), sender, items))
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestWithdrawProcessor) RunPreProcess() *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.RunPreProcess()

	return t
}

func (t *TestWithdrawProcessor) RunProcess() *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.RunProcess()

	return t
}

func (t *TestWithdrawProcessor) IsValid() *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.IsValid()

	return t
}

func (t *TestWithdrawProcessor) Decode(fileName string) *TestWithdrawProcessor {
	t.BaseTestOperationProcessorWithItem.Decode(fileName)

	return t
}
