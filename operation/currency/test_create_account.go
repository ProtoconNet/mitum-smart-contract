package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestCreateAccountProcessor struct {
	*test.BaseTestOperationProcessorWithItem[CreateAccount, CreateAccountItem]
}

func NewTestCreateAccountProcessor(
	tp *test.TestProcessor,
) TestCreateAccountProcessor {
	t := test.NewBaseTestOperationProcessorWithItem[CreateAccount, CreateAccountItem](tp)
	return TestCreateAccountProcessor{&t}
}

func (t *TestCreateAccountProcessor) Create() *TestCreateAccountProcessor {
	t.Opr, _ = NewCreateAccountProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)

	return t
}

func (t *TestCreateAccountProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestCreateAccountProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetAmount(am, cid, target)

	return t
}

func (t *TestCreateAccountProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestCreateAccountProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestCreateAccountProcessor) LoadOperation(fileName string) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.LoadOperation(fileName)

	return t
}

func (t *TestCreateAccountProcessor) Print(fileName string) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.Print(fileName)

	return t
}

func (t *TestCreateAccountProcessor) MakeItem(
	target test.Account, amounts []types.Amount, targetItems []CreateAccountItem,
) *TestCreateAccountProcessor {
	item := NewCreateAccountItemMultiAmounts(target.Keys(), amounts)
	test.UpdateSlice[CreateAccountItem](item, targetItems)

	return t
}

func (t *TestCreateAccountProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, items []CreateAccountItem,
) *TestCreateAccountProcessor {
	op, _ := NewCreateAccount(NewCreateAccountFact([]byte("token"), sender, items))
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestCreateAccountProcessor) RunPreProcess() *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.RunPreProcess()

	return t
}

func (t *TestCreateAccountProcessor) RunProcess() *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.RunProcess()

	return t
}

func (t *TestCreateAccountProcessor) IsValid() *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.IsValid()

	return t
}

func (t *TestCreateAccountProcessor) Decode(fileName string) *TestCreateAccountProcessor {
	t.BaseTestOperationProcessorWithItem.Decode(fileName)

	return t
}
