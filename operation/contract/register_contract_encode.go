package contract

import (
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
)

func (fact *RegisterContractFact) unpack(
	enc encoder.Encoder,
	sa, ta, cc string, cd map[string]string, cid string,
) error {
	fact.currency = ctypes.CurrencyID(cid)

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
	fact.code = cc
	if cd == nil {
		cd = make(map[string]string)
	}
	fact.callData = cd

	return nil
}
