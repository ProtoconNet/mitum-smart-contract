package extension

import (
	"github.com/ProtoconNet/mitum-currency/v2/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseCreateContractAccountsItem struct {
	hint.BaseHinter
	keys        base.AccountKeys
	amounts     []base.Amount
	addressType hint.Type
}

func NewBaseCreateContractAccountsItem(ht hint.Hint, keys base.AccountKeys, amounts []base.Amount, addrType hint.Type) BaseCreateContractAccountsItem {
	return BaseCreateContractAccountsItem{
		BaseHinter:  hint.NewBaseHinter(ht),
		keys:        keys,
		amounts:     amounts,
		addressType: addrType,
	}
}

func (it BaseCreateContractAccountsItem) Bytes() []byte {
	length := 2
	bs := make([][]byte, len(it.amounts)+length)
	bs[0] = it.keys.Bytes()
	bs[1] = it.addressType.Bytes()
	for i := range it.amounts {
		bs[i+length] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseCreateContractAccountsItem) IsValid([]byte) error {
	if n := len(it.amounts); n == 0 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	if err := util.CheckIsValiders(nil, false, it.BaseHinter, it.keys, it.addressType); err != nil {
		return err
	}

	if it.addressType != base.AddressHint.Type() && it.addressType != base.EthAddressHint.Type() {
		return util.ErrInvalid.Errorf("invalid AddressHint")
	}

	founds := map[base.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicate currency found, %q", am.Currency())
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

func (it BaseCreateContractAccountsItem) Keys() base.AccountKeys {
	return it.keys
}

func (it BaseCreateContractAccountsItem) Address() (mitumbase.Address, error) {
	return base.NewAddressFromKeys(it.keys)
}

func (it BaseCreateContractAccountsItem) AddressType() hint.Type {
	return it.addressType
}

func (it BaseCreateContractAccountsItem) Amounts() []base.Amount {
	return it.amounts
}

func (it BaseCreateContractAccountsItem) Rebuild() CreateContractAccountsItem {
	ams := make([]base.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
