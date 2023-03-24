package currency

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
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
		return util.ErrInvalid.Errorf("invalid mitum currency address: %w", err)
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
	if err := keys.IsValid(nil); err != nil {
		return EthAddress{}, err
	}

	k, ok := keys.(BaseAccountKeys)
	if !ok {
		return EthAddress{}, errors.Errorf("expected BaseAccountKeys, not %T", keys)
	}
	h, err := k.GenerateKeccakHash()
	if err != nil {
		return EthAddress{}, err
	}
	v, ok := h.(valuehash.L32)
	if !ok {
		return EthAddress{}, errors.Errorf("expected valuehash.L32, not %T", h)
	}

	return NewEthAddress(hex.EncodeToString(v[12:])), nil
}

func (ca EthAddress) IsValid([]byte) error {
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
