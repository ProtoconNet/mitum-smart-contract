package common

import (
	"context"

	"golang.org/x/exp/slices"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

type BaseOperation struct {
	h     util.Hash
	fact  base.Fact
	signs []base.Sign
	hint.BaseHinter
}

func NewBaseOperation(ht hint.Hint, fact base.Fact) BaseOperation {
	return BaseOperation{
		BaseHinter: hint.NewBaseHinter(ht),
		fact:       fact,
	}
}

func (op BaseOperation) Hash() util.Hash {
	return op.h
}

func (op *BaseOperation) SetHash(h util.Hash) {
	op.h = h
}

func (op BaseOperation) Signs() []base.Sign {
	return op.signs
}

func (op BaseOperation) Fact() base.Fact {
	return op.fact
}

func (op *BaseOperation) SetFact(fact base.Fact) {
	op.fact = fact
}

func (op BaseOperation) HashBytes() []byte {
	bs := make([]util.Byter, len(op.signs)+1)
	bs[0] = op.fact.Hash()

	for i := range op.signs {
		bs[i+1] = op.signs[i]
	}

	return util.ConcatByters(bs...)
}

func (op BaseOperation) IsValid(networkID []byte) error {
	if len(op.signs) < 1 {
		return ErrOperationInvalid.Wrap(errors.Errorf("Empty signs"))
	}

	if err := util.CheckIsValiders(networkID, false, op.h); err != nil {
		return ErrOperationInvalid.Wrap(err)
	}

	sfs := op.Signs()
	var duplicatederr error

	switch duplicated := util.IsDuplicatedSlice(sfs, func(i base.Sign) (bool, string) {
		if i == nil {
			return true, ""
		}

		s, ok := i.(base.Sign)
		if !ok {
			duplicatederr = ErrTypeMismatch.Wrap(errors.Errorf("expected Sign got %T", i))
		}

		return duplicatederr == nil, s.Signer().String()
	}); {
	case duplicatederr != nil:
		return ErrOperationInvalid.Wrap(ErrSignInvalid.Wrap(duplicatederr))
	case duplicated:
		return ErrOperationInvalid.Wrap(ErrSignInvalid.Wrap(errors.Errorf("duplicated signs found")))
	}

	if err := IsValidSignFact(op, networkID); err != nil {
		return ErrOperationInvalid.Wrap(err)
	}

	if !op.h.Equal(op.hash()) {
		return ErrOperationInvalid.Wrap(errors.Errorf("hash does not match"))
	}

	return nil
}

func (op *BaseOperation) Sign(priv base.Privatekey, networkID base.NetworkID) error {
	switch index, sign, err := op.sign(priv, networkID); {
	case err != nil:
		return err
	case index < 0:
		op.signs = append(op.signs, sign)
	default:
		op.signs[index] = sign
	}

	op.h = op.hash()

	return nil
}

func (op *BaseOperation) sign(priv base.Privatekey, networkID base.NetworkID) (found int, sign base.BaseSign, _ error) {
	e := util.StringError("sign BaseOperation")

	found = -1

	for i := range op.signs {
		s := op.signs[i]
		if s == nil {
			continue
		}

		if s.Signer().Equal(priv.Publickey()) {
			found = i

			break
		}
	}

	newsign, err := base.NewBaseSignFromFact(priv, networkID, op.fact)
	if err != nil {
		return found, sign, e.Wrap(err)
	}

	return found, newsign, nil
}

func (BaseOperation) PreProcess(ctx context.Context, _ base.GetStateFunc) (
	context.Context, base.OperationProcessReasonError, error,
) {
	return ctx, nil, errors.WithStack(util.ErrNotImplemented)
}

func (BaseOperation) Process(context.Context, base.GetStateFunc) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	return nil, nil, errors.WithStack(util.ErrNotImplemented)
}

func (op BaseOperation) hash() util.Hash {
	return valuehash.NewSHA256(op.HashBytes())
}

func IsValidOperationFact(fact base.Fact, networkID []byte) error {
	if err := util.CheckIsValiders(networkID, false,
		fact.Hash(),
	); err != nil {
		return err
	}

	switch l := len(fact.Token()); {
	case l < 1:
		return errors.Errorf("Operation has empty token")
	case l > base.MaxTokenSize:
		return errors.Errorf("Operation token size too large: %d > %d", l, base.MaxTokenSize)
	}

	hg, ok := fact.(HashGenerator)
	if !ok {
		return nil
	}

	if !fact.Hash().Equal(hg.GenerateHash()) {
		return errors.Errorf("Wrong Fact hash")
	}

	return nil
}

type BaseNodeOperation struct {
	BaseOperation
}

func NewBaseNodeOperation(ht hint.Hint, fact base.Fact) BaseNodeOperation {
	return BaseNodeOperation{
		BaseOperation: NewBaseOperation(ht, fact),
	}
}

func (op BaseNodeOperation) IsValid(networkID []byte) error {
	if err := op.BaseOperation.IsValid(networkID); err != nil {
		return ErrNodeOperationInvalid.Wrap(err)
	}

	sfs := op.Signs()

	var duplicatederr error

	switch duplicated := util.IsDuplicatedSlice(sfs, func(i base.Sign) (bool, string) {
		if i == nil {
			return true, ""
		}

		ns, ok := i.(base.NodeSign)
		if !ok {
			duplicatederr = errors.Errorf("expected NodeSign got %T", i)
		}

		return duplicatederr == nil, ns.Node().String()
	}); {
	case duplicatederr != nil:
		return ErrNodeOperationInvalid.Wrap(duplicatederr)
	case duplicated:
		return ErrNodeOperationInvalid.Wrap(errors.Errorf("Duplicated signs found"))
	}

	for i := range sfs {
		if _, ok := sfs[i].(base.NodeSign); !ok {
			return ErrNodeOperationInvalid.Wrap(errors.Errorf("expected NodeSign got %T", sfs[i]))
		}
	}

	return nil
}

func (op *BaseNodeOperation) NodeSign(priv base.Privatekey, networkID base.NetworkID, node base.Address) error {
	found := -1

	for i := range op.signs {
		s := op.signs[i].(base.NodeSign) //nolint:forcetypeassert //...
		if s == nil {
			continue
		}

		if s.Node().Equal(node) {
			found = i

			break
		}
	}

	ns, err := base.NewBaseNodeSignFromFact(node, priv, networkID, op.fact)
	if err != nil {
		return err
	}

	switch {
	case found < 0:
		op.signs = append(op.signs, ns)
	default:
		op.signs[found] = ns
	}

	op.h = op.hash()

	return nil
}

func (op *BaseNodeOperation) SetNodeSigns(signs []base.NodeSign) error {
	if duplicated := util.IsDuplicatedSlice(signs, func(i base.NodeSign) (bool, string) {
		if i == nil {
			return true, ""
		}

		return true, i.Node().String()
	}); duplicated {
		return errors.Errorf("Duplicated signs found")
	}

	op.signs = make([]base.Sign, len(signs))
	for i := range signs {
		op.signs[i] = signs[i]
	}

	op.h = op.hash()

	return nil
}

func (op *BaseNodeOperation) AddNodeSigns(signs []base.NodeSign) (added bool, _ error) {
	updates := util.FilterSlice(signs, func(sign base.NodeSign) bool {
		return slices.IndexFunc(op.signs, func(s base.Sign) bool {
			nodesign, ok := s.(base.NodeSign)
			if !ok {
				return false
			}

			return sign.Node().Equal(nodesign.Node())
		}) < 0
	})

	if len(updates) < 1 {
		return false, nil
	}

	mergedsigns := make([]base.Sign, len(op.signs)+len(updates))
	copy(mergedsigns, op.signs)

	for i := range updates {
		mergedsigns[len(op.signs)+i] = updates[i]
	}

	op.signs = mergedsigns
	op.h = op.hash()

	return true, nil
}

func (op BaseNodeOperation) NodeSigns() []base.NodeSign {
	ss := op.Signs()
	signs := make([]base.NodeSign, len(ss))

	for i := range ss {
		signs[i] = ss[i].(base.NodeSign) //nolint:forcetypeassert //...
	}

	return signs
}

type HashGenerator interface {
	GenerateHash() util.Hash
}
