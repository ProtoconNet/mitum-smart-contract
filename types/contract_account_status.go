package types // nolint: dupl, revive

import (
	"regexp"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var ContractAccountStatusHint = hint.MustNewHint("mitum-currency-contract-account-status-v0.0.1")

type ContractAccountStatus struct {
	hint.BaseHinter
	owner    base.Address
	isActive bool
}

func NewContractAccountStatus(owner base.Address, isActive bool) ContractAccountStatus {
	us := ContractAccountStatus{
		BaseHinter: hint.NewBaseHinter(ContractAccountStatusHint),
		owner:      owner,
		isActive:   isActive,
	}
	return us
}

func (cs ContractAccountStatus) Bytes() []byte {
	var v int8
	if cs.isActive {
		v = 1
	}

	return util.ConcatBytesSlice(cs.owner.Bytes(), []byte{byte(v)})
}

func (cs ContractAccountStatus) Hash() util.Hash {
	return cs.GenerateHash()
}

func (cs ContractAccountStatus) GenerateHash() util.Hash {
	return valuehash.NewSHA256(cs.Bytes())
}

func (cs ContractAccountStatus) IsValid([]byte) error { // nolint:revive
	return nil
}

func (cs ContractAccountStatus) Owner() base.Address { // nolint:revive
	return cs.owner
}

func (cs ContractAccountStatus) SetOwner(a base.Address) (ContractAccountStatus, error) { // nolint:revive
	err := a.IsValid(nil)
	if err != nil {
		return ContractAccountStatus{}, err
	}

	cs.owner = a

	return cs, nil
}

func (cs ContractAccountStatus) IsActive() bool { // nolint:revive
	return cs.isActive
}

func (cs ContractAccountStatus) SetIsActive(b bool) ContractAccountStatus { // nolint:revive
	cs.isActive = b
	return cs
}

func (cs ContractAccountStatus) Equal(b ContractAccountStatus) bool {
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
