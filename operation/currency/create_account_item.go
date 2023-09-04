package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseCreateAccountItem struct {
	hint.BaseHinter
	keys        types.AccountKeys
	amounts     []types.Amount
	addressType hint.Type
}

func NewBaseCreateAccountItem(ht hint.Hint, keys types.AccountKeys, amounts []types.Amount, addrHint hint.Type) BaseCreateAccountItem {
	return BaseCreateAccountItem{
		BaseHinter:  hint.NewBaseHinter(ht),
		keys:        keys,
		amounts:     amounts,
		addressType: addrHint,
	}
}

func (it BaseCreateAccountItem) Bytes() []byte {
	bs := make([][]byte, len(it.amounts)+2)
	bs[0] = it.keys.Bytes()
	bs[1] = it.addressType.Bytes()
	for i := range it.amounts {
		bs[i+2] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseCreateAccountItem) IsValid([]byte) error {
	if n := len(it.amounts); n == 0 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	if err := util.CheckIsValiders(nil, false, it.BaseHinter, it.keys); err != nil {
		return err
	}

	if it.addressType != types.AddressHint.Type() && it.addressType != types.EthAddressHint.Type() {
		return util.ErrInvalid.Errorf("invalid AddressHint")
	}

	founds := map[types.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicated currency found, %v", am.Currency())
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

func (it BaseCreateAccountItem) Keys() types.AccountKeys {
	return it.keys
}

func (it BaseCreateAccountItem) Address() (base.Address, error) {
	if it.addressType == types.AddressHint.Type() {
		return types.NewAddressFromKeys(it.keys)
	} else if it.addressType == types.EthAddressHint.Type() {
		return types.NewEthAddressFromKeys(it.keys)
	}
	return nil, util.ErrInvalid.Errorf("invalid address hint")
}

func (it BaseCreateAccountItem) AddressType() hint.Type {
	return it.addressType
}

func (it BaseCreateAccountItem) Amounts() []types.Amount {
	return it.amounts
}

func (it BaseCreateAccountItem) Rebuild() CreateAccountItem {
	ams := make([]types.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
