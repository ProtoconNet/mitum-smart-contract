package types // nolint: dupl, revive

import (
	"bytes"
	"regexp"
	"sort"

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
	handlers []base.Address
}

func NewContractAccountStatus(owner base.Address, handlers []base.Address) ContractAccountStatus {
	sort.Slice(handlers, func(i, j int) bool {
		return bytes.Compare(handlers[i].Bytes(), handlers[j].Bytes()) < 0
	})

	us := ContractAccountStatus{
		BaseHinter: hint.NewBaseHinter(ContractAccountStatusHint),
		owner:      owner,
		isActive:   false,
		handlers:   handlers,
	}
	return us
}

func (cs ContractAccountStatus) Bytes() []byte {
	var v int8
	if cs.isActive {
		v = 1
	}
	handlers := make([][]byte, len(cs.handlers))
	for i := range cs.handlers {
		handlers[i] = cs.handlers[i].Bytes()
	}

	return util.ConcatBytesSlice(cs.owner.Bytes(), []byte{byte(v)}, util.ConcatBytesSlice(handlers...))
}

func (cs ContractAccountStatus) Hash() util.Hash {
	return cs.GenerateHash()
}

func (cs ContractAccountStatus) GenerateHash() util.Hash {
	return valuehash.NewSHA256(cs.Bytes())
}

func (cs ContractAccountStatus) IsValid([]byte) error { // nolint:revive
	if err := util.CheckIsValiders(nil, false,
		cs.BaseHinter,
		cs.owner,
	); err != nil {
		return err
	}

	return nil
}

func (cs ContractAccountStatus) Owner() base.Address { // nolint:revive
	return cs.owner
}

func (cs *ContractAccountStatus) SetOwner(a base.Address) error { // nolint:revive
	err := a.IsValid(nil)
	if err != nil {
		return err
	}

	cs.owner = a

	return nil
}

func (cs ContractAccountStatus) Handlers() []base.Address { // nolint:revive
	return cs.handlers
}

func (cs *ContractAccountStatus) SetHandlers(handlers []base.Address) error {
	sort.Slice(handlers, func(i, j int) bool {
		return bytes.Compare(handlers[i].Bytes(), handlers[j].Bytes()) < 0
	})

	for i := range handlers {
		err := handlers[i].IsValid(nil)
		if err != nil {
			return err
		}
	}

	cs.handlers = handlers

	return nil
}

func (cs ContractAccountStatus) IsHandler(ad base.Address) bool { // nolint:revive
	for i := range cs.Handlers() {
		if ad.Equal(cs.Handlers()[i]) {
			return true
		}
	}
	return false
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

	for i := range cs.handlers {
		if !cs.handlers[i].Equal(b.handlers[i]) {
			return false
		}
	}

	return true
}

var (
	MinLengthContractID = 3
	MaxLengthContractID = 50
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
