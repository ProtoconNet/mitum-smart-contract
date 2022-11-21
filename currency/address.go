package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	AddressHint       = hint.MustNewHint("mca-v0.0.1")
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
		return util.ErrInvalid.Errorf("invalid mitum currency address: %w", err)
	}

	return nil
}

type Addresses interface {
	Addresses() ([]base.Address, error)
}

func ZeroAddress(cid CurrencyID) Address {
	return NewAddress(cid.String() + ZeroAddressSuffix)
}
