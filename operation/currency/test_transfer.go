package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestTransferProcessor struct {
	*test.BaseTestOperationProcessorWithItem[Transfer, TransferItem]
}

func NewTestTransferProcessor(
	tp *test.TestProcessor,
) TestTransferProcessor {
	t := test.NewBaseTestOperationProcessorWithItem[Transfer, TransferItem](tp)
	return TestTransferProcessor{&t}
}

func (t *TestTransferProcessor) Create() *TestTransferProcessor {
	t.Opr, _ = NewTransferProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)

	return t
}

func (t *TestTransferProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestTransferProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.SetAmount(am, cid, target)

	return t
}

func (t *TestTransferProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestTransferProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestTransferProcessor) LoadOperation(fileName string) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.LoadOperation(fileName)

	return t
}

func (t *TestTransferProcessor) Print(fileName string) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.Print(fileName)

	return t
}

func (t *TestTransferProcessor) MakeItem(
	receiver test.Account, amounts []types.Amount, targetItems []TransferItem,
) *TestTransferProcessor {
	item := NewTransferItemMultiAmounts(receiver.Address(), amounts)
	test.UpdateSlice[TransferItem](item, targetItems)

	return t
}

func (t *TestTransferProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, items []TransferItem,
) *TestTransferProcessor {
	op, _ := NewTransfer(NewTransferFact([]byte("token"), sender, items))
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestTransferProcessor) RunPreProcess() *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.RunPreProcess()

	return t
}

func (t *TestTransferProcessor) RunProcess() *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.RunProcess()

	return t
}

func (t *TestTransferProcessor) IsValid() *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.IsValid()

	return t
}

func (t *TestTransferProcessor) Decode(fileName string) *TestTransferProcessor {
	t.BaseTestOperationProcessorWithItem.Decode(fileName)

	return t
}
