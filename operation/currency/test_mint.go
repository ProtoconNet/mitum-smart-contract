package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/operation/test"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

type TestMintProcessor struct {
	*test.BaseTestOperationProcessorWithItem[Mint, MintItem]
}

func NewTestMintProcessor(
	tp *test.TestProcessor,
) TestMintProcessor {
	t := test.NewBaseTestOperationProcessorWithItem[Mint, MintItem](tp)

	return TestMintProcessor{&t}
}

func (t *TestMintProcessor) Create() *TestMintProcessor {
	t.Opr, _ = NewMintProcessor(base.MaxThreshold)(
		base.GenesisHeight,
		t.GetStateFunc,
		nil, nil,
	)

	return t
}

func (t *TestMintProcessor) SetCurrency(
	cid string, am int64, addr base.Address, target []types.CurrencyID, instate bool) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.SetCurrency(cid, am, addr, target, instate)

	return t
}

func (t *TestMintProcessor) SetAmount(am int64, cid types.CurrencyID, target []types.Amount) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.SetAmount(am, cid, target)

	return t
}

func (t *TestMintProcessor) SetContractAccount(
	owner base.Address, priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool,
) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.SetContractAccount(owner, priv, amount, cid, target, inState)

	return t
}

func (t *TestMintProcessor) SetAccount(
	priv string, amount int64, cid types.CurrencyID, target []test.Account, inState bool) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.SetAccount(priv, amount, cid, target, inState)

	return t
}

func (t *TestMintProcessor) LoadOperation(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.LoadOperation(fileName)

	return t
}

func (t *TestMintProcessor) Print(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.Print(fileName)

	return t
}

func (t *TestMintProcessor) MakeItem(
	receiver test.Account, amount types.Amount, targetItems []MintItem,
) *TestMintProcessor {
	item := NewMintItem(receiver.Address(), amount)
	test.UpdateSlice[MintItem](item, targetItems)

	return t
}

func (t *TestMintProcessor) MakeOperation(items []MintItem,
) *TestMintProcessor {
	//t.MockGetter.On("Get", mock.Anything).Return(nil, false, nil)

	op, _ := NewMint(NewMintFact([]byte("token"), items))
	_ = op.NodeSign(t.NodePriv, t.NetworkID, t.NodeAddr)
	t.Op = op

	return t
}

func (t *TestMintProcessor) RunPreProcess() *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.RunPreProcess()

	return t
}

func (t *TestMintProcessor) RunProcess() *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.RunProcess()

	return t
}

func (t *TestMintProcessor) IsValid() *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.IsValid()

	return t
}

func (t *TestMintProcessor) Decode(fileName string) *TestMintProcessor {
	t.BaseTestOperationProcessorWithItem.Decode(fileName)

	return t
}
