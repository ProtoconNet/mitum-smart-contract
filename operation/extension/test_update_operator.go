package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum2/util/encoder"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestUpdateOperatorProcessor struct {
	*test.BaseTestOperationProcessorNoItem[UpdateOperator]
}

func NewTestUpdateOperatorProcessor(encs *encoder.Encoders) TestUpdateOperatorProcessor {
	t := test.NewBaseTestOperationProcessorNoItem[UpdateOperator](encs)
	return TestUpdateOperatorProcessor{&t}
}

func (t *TestUpdateOperatorProcessor) Create() *TestUpdateOperatorProcessor {
	t.Opr, _ = NewUpdateOperatorProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)
	return t
}

func (t *TestUpdateOperatorProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestUpdateOperatorProcessor) SetAmount(
	am int64, cid types.CurrencyID, target []types.Amount,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.SetAmount(am, cid, target)

	return t
}

func (t *TestUpdateOperatorProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateOperatorProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateOperatorProcessor) LoadOperation(fileName string,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.LoadOperation(fileName)

	return t
}

func (t *TestUpdateOperatorProcessor) Print(fileName string,
) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.Print(fileName)

	return t
}

func (t *TestUpdateOperatorProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, contract base.Address, operators []test.Account, currency types.CurrencyID,
) *TestUpdateOperatorProcessor {
	var oprs []base.Address
	for _, operator := range operators {
		oprs = append(oprs, operator.Address())
	}

	op, _ := NewUpdateOperator(
		NewUpdateOperatorFact(
			[]byte("token"), sender, contract, oprs, currency,
		),
	)
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestUpdateOperatorProcessor) RunPreProcess() *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.RunPreProcess()

	return t
}

func (t *TestUpdateOperatorProcessor) RunProcess() *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.RunProcess()

	return t
}

func (t *TestUpdateOperatorProcessor) IsValid() *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.IsValid()

	return t
}

func (t *TestUpdateOperatorProcessor) Decode(fileName string) *TestUpdateOperatorProcessor {
	t.BaseTestOperationProcessorNoItem.Decode(fileName)

	return t
}
