package currency

import (
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseCreateAccountsItem struct {
	hint.BaseHinter
	keys        base3.AccountKeys
	amounts     []base3.Amount
	addressType hint.Type
}

func NewBaseCreateAccountsItem(ht hint.Hint, keys base3.AccountKeys, amounts []base3.Amount, addrHint hint.Type) BaseCreateAccountsItem {
	return BaseCreateAccountsItem{
		BaseHinter:  hint.NewBaseHinter(ht),
		keys:        keys,
		amounts:     amounts,
		addressType: addrHint,
	}
}

func (it BaseCreateAccountsItem) Bytes() []byte {
	bs := make([][]byte, len(it.amounts)+2)
	bs[0] = it.keys.Bytes()
	bs[1] = it.addressType.Bytes()
	for i := range it.amounts {
		bs[i+2] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseCreateAccountsItem) IsValid([]byte) error {
	if n := len(it.amounts); n == 0 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	if err := util.CheckIsValiders(nil, false, it.BaseHinter, it.keys); err != nil {
		return err
	}

	if it.addressType != base3.AddressHint.Type() && it.addressType != base3.EthAddressHint.Type() {
		return util.ErrInvalid.Errorf("invalid AddressHint")
	}

	founds := map[base3.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicated currency found, %q", am.Currency())
		}
		founds[am.Currency()] = struct{}{}

		if err := am.IsValid(nil); err != nil {
			return err
		} else if !am.Big().OverZero() {
			return util.ErrInvalid.Errorf("amount should be over zero")
		}
	}

	return nil
}

func (it BaseCreateAccountsItem) Keys() base3.AccountKeys {
	return it.keys
}

func (it BaseCreateAccountsItem) Address() (base.Address, error) {
	if it.addressType == base3.AddressHint.Type() {
		return base3.NewAddressFromKeys(it.keys)
	} else if it.addressType == base3.EthAddressHint.Type() {
		return base3.NewEthAddressFromKeys(it.keys)
	}
	return nil, util.ErrInvalid.Errorf("invalid address hint")
}

func (it BaseCreateAccountsItem) AddressType() hint.Type {
	return it.addressType
}

func (it BaseCreateAccountsItem) Amounts() []base3.Amount {
	return it.amounts
}

func (it BaseCreateAccountsItem) Rebuild() CreateAccountsItem {
	ams := make([]base3.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
