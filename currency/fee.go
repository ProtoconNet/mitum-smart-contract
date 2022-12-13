package currency

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	FeeOperationFactHint = hint.MustNewHint("mitum-currency-fee-operation-fact-v0.0.1")
	FeeOperationHint     = hint.MustNewHint("mitum-currency-fee-operation-v0.0.1")
)

type FeeOperationFact struct {
	base.BaseFact
	amounts []Amount
}

func NewFeeOperationFact(height base.Height, ams map[CurrencyID]Big) FeeOperationFact {
	amounts := make([]Amount, len(ams))
	var i int
	for cid := range ams {
		amounts[i] = NewAmount(ams[cid], cid)
		i++
	}

	// TODO replace random bytes with height
	fact := FeeOperationFact{
		BaseFact: base.NewBaseFact(FeeOperationFactHint, height.Bytes()),
		amounts:  amounts,
	}
	fact.SetHash(valuehash.NewSHA256(fact.Bytes()))

	return fact
}

func (fact FeeOperationFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact FeeOperationFact) Bytes() []byte {
	bs := make([][]byte, len(fact.amounts)+1)
	bs[0] = fact.Token()

	for i := range fact.amounts {
		bs[i+1] = fact.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (fact FeeOperationFact) IsValid([]byte) error {
	if len(fact.Token()) < 1 {
		return util.ErrInvalid.Errorf("empty token for FeeOperationFact")
	}

	if err := util.CheckIsValiders(nil, false, fact.Hash()); err != nil {
		return err
	}

	for i := range fact.amounts {
		if err := fact.amounts[i].IsValid(nil); err != nil {
			return err
		}
	}

	return nil
}

func (fact FeeOperationFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact FeeOperationFact) Amounts() []Amount {
	return fact.amounts
}

type FeeOperation struct {
	BaseOperation
}

func NewFeeOperation(fact FeeOperationFact) (FeeOperation, error) {
	return FeeOperation{BaseOperation: NewBaseOperationFromFact(FeeOperationHint, fact, "")}, nil
}

func (op *FeeOperation) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}

func (FeeOperation) Process(
	func(key string) (base.State, bool, error),
	func(util.Hash, ...base.State) error,
) error {
	return nil
}

type FeeOperationProcessor struct {
	*base.BaseOperationProcessor
}

func NewFeeOperationProcessor(
	height base.Height,
	getStateFunc base.GetStateFunc,
) (base.OperationProcessor, error) {
	e := util.StringErrorFunc("failed to create new FeeOperationProcessor")

	b, err := base.NewBaseOperationProcessor(
		height, getStateFunc, nil, nil)
	if err != nil {
		return nil, e(err, "")
	}
	return &FeeOperationProcessor{
		BaseOperationProcessor: b,
	}, nil
}

func (opp *FeeOperationProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	return ctx, nil, nil
}

func (opp *FeeOperationProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(FeeOperationFact)
	if !ok {
		return nil, nil, errors.Errorf("expected FeeOperationFact, not %T", op.Fact())
	}

	sts := make([]base.StateMergeValue, len(fact.amounts))
	for i := range fact.amounts {
		am := fact.amounts[i]

		policy, err := existsCurrencyPolicy(am.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		if policy.feeer.Receiver() == nil {
			continue
		}

		if err := checkExistsState(StateKeyAccount(policy.feeer.Receiver()), getStateFunc); err != nil {
			return nil, nil, err
		} else if st, _, err := getStateFunc(StateKeyBalance(policy.feeer.Receiver(), am.Currency())); err != nil {
			return nil, nil, err
		} else {
			v, ok := st.Value().(BalanceStateValue)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError("invalid BalanceState value found, %T", st.Value()), nil
			}
			sts[i] = NewBalanceStateMergeValue(st.Key(), NewBalanceStateValue(v.Amount.WithBig(v.Amount.big.Add(am.big))))
		}
	}

	return sts, nil, nil
}
