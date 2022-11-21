package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

type BaseOperation struct {
	base.BaseOperation
	Memo string
}

func NewBaseOperationFromFact(
	ht hint.Hint, fact base.Fact, memo string,
) BaseOperation {
	bo := base.NewBaseOperation(ht, fact)
	op := BaseOperation{BaseOperation: bo, Memo: memo}

	return op
}

func (op *BaseOperation) Sign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.BaseOperation.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}

func (op BaseOperation) IsValid(networkID []byte) error {
	if err := op.BaseOperation.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := IsValidMemo(op.Memo); err != nil {
		return err
	}

	return op.BaseOperation.IsValid(networkID)
}

func (op BaseOperation) GenerateHash() util.Hash {
	bs := make([][]byte, len(op.Signs())+1)
	for i := range op.Signs() {
		bs[i] = op.Signs()[i].Bytes()
	}

	bs[len(bs)-1] = []byte(op.Memo)

	e := util.ConcatBytesSlice(op.Fact().Hash().Bytes(), util.ConcatBytesSlice(bs...))

	return valuehash.NewSHA256(e)
}

func operationHinter(ht hint.Hint) BaseOperation {
	return BaseOperation{BaseOperation: base.BaseOperation{BaseHinter: hint.NewBaseHinter(ht)}}

}

func IsValidOperationFact(fact base.Fact, networkID []byte) error {
	if err := util.CheckIsValiders(networkID, false,
		fact.Hash(),
	); err != nil {
		return err
	}

	switch l := len(fact.Token()); {
	case l < 1:
		return util.ErrInvalid.Errorf("Operation has empty token")
	case l > base.MaxTokenSize:
		return util.ErrInvalid.Errorf("Operation token size too large: %d > %d", l, base.MaxTokenSize)
	}

	hg, ok := fact.(HashGenerator)
	if !ok {
		return nil
	}

	if !fact.Hash().Equal(hg.GenerateHash()) {
		return util.ErrInvalid.Errorf("wrong Fact hash")
	}

	return nil
}

type HashGenerator interface {
	GenerateHash() util.Hash
}
