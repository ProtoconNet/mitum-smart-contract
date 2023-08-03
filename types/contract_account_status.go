package types // nolint: dupl, revive

import (
	"regexp"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var ContractAccountHint = hint.MustNewHint("mitum-currency-contract-account-status-v0.0.1")

type ContractAccount struct {
	hint.BaseHinter
	owner    base.Address
	isActive bool
}

func NewContractAccount(owner base.Address, isActive bool) ContractAccount {
	us := ContractAccount{
		BaseHinter: hint.NewBaseHinter(ContractAccountHint),
		owner:      owner,
		isActive:   isActive,
	}
	return us
}

func (cs ContractAccount) Bytes() []byte {
	var v int8
	if cs.isActive {
		v = 1
	}

	return util.ConcatBytesSlice(cs.owner.Bytes(), []byte{byte(v)})
}

func (cs ContractAccount) Hash() util.Hash {
	return cs.GenerateHash()
}

func (cs ContractAccount) GenerateHash() util.Hash {
	return valuehash.NewSHA256(cs.Bytes())
}

func (cs ContractAccount) IsValid([]byte) error { // nolint:revive
	return nil
}

func (cs ContractAccount) Owner() base.Address { // nolint:revive
	return cs.owner
}

func (cs ContractAccount) SetOwner(a base.Address) (ContractAccount, error) { // nolint:revive
	err := a.IsValid(nil)
	if err != nil {
		return ContractAccount{}, err
	}

	cs.owner = a

	return cs, nil
}

func (cs ContractAccount) IsActive() bool { // nolint:revive
	return cs.isActive
}

func (cs ContractAccount) SetIsActive(b bool) ContractAccount { // nolint:revive
	cs.isActive = b
	return cs
}

func (cs ContractAccount) Equal(b ContractAccount) bool {
	if cs.isActive != b.isActive {
		return false
	}
	if !cs.owner.Equal(b.owner) {
		return false
	}

	return true
}

var (
	MinLengthContractID = 3
	MaxLengthContractID = 10
	REContractIDExp     = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-_\.\!\$\*\@]*[A-Z0-9]$`)
)

type ContractID string

func (cid ContractID) Bytes() []byte {
	return []byte(cid)
}

func (cid ContractID) String() string {
	return string(cid)
}

func (cid ContractID) IsValid([]byte) error {
	if l := len(cid); l < MinLengthContractID || l > MaxLengthContractID {
		return util.ErrInvalid.Errorf(
			"invalid length of contract id, %d <= %d <= %d", MinLengthContractID, l, MaxLengthContractID)
	}
	if !REContractIDExp.Match([]byte(cid)) {
		return util.ErrInvalid.Errorf("wrong contract id, %q", cid)
	}

	return nil
}
