package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	AddressHint       = hint.MustNewHint("mca-v0.0.1")
	EthAddressHint    = hint.MustNewHint("eca-v0.0.1")
	ZeroAddressSuffix = "-X"
)

type Address struct {
	base.BaseStringAddress
}

func NewAddress(s string) Address {
	ca := Address{BaseStringAddress: base.NewBaseStringAddressWithHint(AddressHint, s)}

	return ca
}

func NewAddressFromKeys(keys AccountKeys) (Address, error) {
	if err := keys.IsValid(nil); err != nil {
		return Address{}, err
	}

	return NewAddress(keys.Hash().String()), nil
}

func (ca Address) IsValid([]byte) error {
	if err := ca.BaseStringAddress.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}

	return nil
}

type EthAddress struct {
	base.BaseStringAddress
}

func NewEthAddress(s string) EthAddress {
	ca := EthAddress{BaseStringAddress: base.NewBaseStringAddressWithHint(EthAddressHint, s)}

	return ca
}

func NewEthAddressFromKeys(keys AccountKeys) (EthAddress, error) {
	//var b valuehash.L32
	//copy(b[:], keys.Hash().Bytes()[:])
	//
	//return NewEthAddress(hex.EncodeToString(b[12:])), nil
	return NewEthAddress(keys.Hash().String()), nil
}

func (ca EthAddress) IsValid([]byte) error {
	if err := ca.BaseStringAddress.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid mitum currency address: %v", err)
	}

	return nil
}

type Addresses interface {
	Addresses() ([]base.Address, error)
}

func ZeroAddress(cid CurrencyID) Address {
	return NewAddress(cid.String() + ZeroAddressSuffix)
}
