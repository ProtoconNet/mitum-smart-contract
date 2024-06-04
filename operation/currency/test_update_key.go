package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestUpdateKeyProcessor struct {
	*test.BaseTestOperationProcessorNoItem[UpdateKey]
}

func NewTestUpdateKeyProcessor(
	tp *test.TestProcessor,
) TestUpdateKeyProcessor {
	t := test.NewBaseTestOperationProcessorNoItem[UpdateKey](tp)
	return TestUpdateKeyProcessor{&t}
}

func (t *TestUpdateKeyProcessor) Create() *TestUpdateKeyProcessor {
	t.Opr, _ = NewUpdateKeyProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)

	return t
}

func (t *TestUpdateKeyProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestUpdateKeyProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.SetAmount(am, cid, target)

	return t
}

func (t *TestUpdateKeyProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateKeyProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateKeyProcessor) LoadOperation(fileName string) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.LoadOperation(fileName)

	return t
}

func (t *TestUpdateKeyProcessor) Print(fileName string) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.Print(fileName)

	return t
}

func (t *TestUpdateKeyProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, target types.AccountKeys, currency types.CurrencyID,
) *TestUpdateKeyProcessor {
	//t.MockGetter.On("Get", mock.Anything).Return(nil, false, nil)

	op, _ := NewUpdateKey(NewUpdateKeyFact([]byte("token"), sender, target, currency))
	_ = op.Sign(privatekey, t.NetworkID)

	t.Op = op

	return t
}

func (t *TestUpdateKeyProcessor) RunPreProcess() *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.RunPreProcess()

	return t
}

func (t *TestUpdateKeyProcessor) RunProcess() *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.RunProcess()

	return t
}

func (t *TestUpdateKeyProcessor) IsValid() *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.IsValid()

	return t
}

func (t *TestUpdateKeyProcessor) Decode(fileName string) *TestUpdateKeyProcessor {
	t.BaseTestOperationProcessorNoItem.Decode(fileName)

	return t
}
