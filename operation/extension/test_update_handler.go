package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestUpdateHandlerProcessor struct {
	*test.BaseTestOperationProcessorNoItem[UpdateHandler]
}

func NewTestUpdateHandlerProcessor(
	tp *test.TestProcessor,
) TestUpdateHandlerProcessor {
	t := test.NewBaseTestOperationProcessorNoItem[UpdateHandler](tp)
	return TestUpdateHandlerProcessor{&t}
}

func (t *TestUpdateHandlerProcessor) Create() *TestUpdateHandlerProcessor {
	t.Opr, _ = NewUpdateHandlerProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)
	return t
}

func (t *TestUpdateHandlerProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestUpdateHandlerProcessor) SetAmount(
	am int64, cid types.CurrencyID, target []types.Amount,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.SetAmount(am, cid, target)

	return t
}

func (t *TestUpdateHandlerProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateHandlerProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestUpdateHandlerProcessor) LoadOperation(fileName string,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.LoadOperation(fileName)

	return t
}

func (t *TestUpdateHandlerProcessor) Print(fileName string,
) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.Print(fileName)

	return t
}

func (t *TestUpdateHandlerProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, contract base.Address, handlers []test.Account, currency types.CurrencyID,
) *TestUpdateHandlerProcessor {
	var oprs []base.Address
	for _, handler := range handlers {
		oprs = append(oprs, handler.Address())
	}

	op, _ := NewUpdateHandler(
		NewUpdateHandlerFact(
			[]byte("token"), sender, contract, oprs, currency,
		),
	)
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestUpdateHandlerProcessor) RunPreProcess() *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.RunPreProcess()

	return t
}

func (t *TestUpdateHandlerProcessor) RunProcess() *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.RunProcess()

	return t
}

func (t *TestUpdateHandlerProcessor) IsValid() *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.IsValid()

	return t
}

func (t *TestUpdateHandlerProcessor) Decode(fileName string) *TestUpdateHandlerProcessor {
	t.BaseTestOperationProcessorNoItem.Decode(fileName)

	return t
}
