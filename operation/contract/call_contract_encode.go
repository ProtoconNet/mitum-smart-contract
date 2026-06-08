package contract

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CallContractFact) unpackLegacy(
	enc encoder.Encoder,
	sa, ta string, cd map[string]string, cid string,
) error {
	if err := runtime.ValidateContractCallDataLimits("call_data", cd); err != nil {
		return err
	}
	items := normalizeLegacyCallData(cd)

	return fact.unpack(enc, sa, ta, items, cid)
}

func (fact *CallContractFact) unpackItems(
	enc encoder.Encoder,
	sa, ta string, items []CallContractItem, cid string,
) error {
	return fact.unpack(enc, sa, ta, items, cid)
}

func (fact *CallContractFact) unpack(
	enc encoder.Encoder,
	sa, ta string, items []CallContractItem, cid string,
) error {
	fact.currency = types.CurrencyID(cid)

	sender, err := base.DecodeAddress(sa, enc)
	if err != nil {
		return err
	}
	fact.sender = sender
	contract, err := base.DecodeAddress(ta, enc)
	if err != nil {
		return err
	}
	fact.contract = contract
	fact.items = copyCallContractItems(items)

	return nil
}
