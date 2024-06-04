package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestCreateContractAccountProcessor struct {
	*test.BaseTestOperationProcessorWithItem[CreateContractAccount, CreateContractAccountItem]
}

func NewTestCreateContractAccountProcessor(
	tp *test.TestProcessor,
) TestCreateContractAccountProcessor {
	t := test.NewBaseTestOperationProcessorWithItem[CreateContractAccount, CreateContractAccountItem](tp)
	return TestCreateContractAccountProcessor{&t}
}

func (t *TestCreateContractAccountProcessor) Create() *TestCreateContractAccountProcessor {
	t.Opr, _ = NewCreateContractAccountProcessor()(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)
	return t
}

func (t *TestCreateContractAccountProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestCreateContractAccountProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetAmount(am, cid, target)

	return t
}

func (t *TestCreateContractAccountProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestCreateContractAccountProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestCreateContractAccountProcessor) LoadOperation(fileName string) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.LoadOperation(fileName)

	return t
}

func (t *TestCreateContractAccountProcessor) Print(fileName string) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.Print(fileName)

	return t
}

func (t *TestCreateContractAccountProcessor) MakeItem(
	target test.Account, amounts []types.Amount, targetItems []CreateContractAccountItem,
) *TestCreateContractAccountProcessor {
	item := NewCreateContractAccountItemMultiAmounts(target.Keys(), amounts)
	test.UpdateSlice[CreateContractAccountItem](item, targetItems)

	return t
}

func (t *TestCreateContractAccountProcessor) MakeOperation(
	sender base.Address, privatekey base.Privatekey, items []CreateContractAccountItem,
) *TestCreateContractAccountProcessor {
	op, _ := NewCreateContractAccount(NewCreateContractAccountFact([]byte("token"), sender, items))
	_ = op.Sign(privatekey, t.NetworkID)
	t.Op = op

	return t
}

func (t *TestCreateContractAccountProcessor) RunPreProcess() *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.RunPreProcess()

	return t
}

func (t *TestCreateContractAccountProcessor) RunProcess() *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.RunProcess()

	return t
}

func (t *TestCreateContractAccountProcessor) IsValid() *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.IsValid()

	return t
}

func (t *TestCreateContractAccountProcessor) Decode(fileName string) *TestCreateContractAccountProcessor {
	t.BaseTestOperationProcessorWithItem.Decode(fileName)

	return t
}
