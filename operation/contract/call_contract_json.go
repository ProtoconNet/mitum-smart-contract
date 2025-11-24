package contract

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type CallContractFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender   base.Address      `json:"sender"`
	Contract base.Address      `json:"contract"`
	CallData map[string]string `json:"call_data"`
	Currency types.CurrencyID  `json:"currency"`
}

func (fact CallContractFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CallContractFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		CallData:              fact.callData,
		Currency:              fact.currency,
	})
}

type CallContractFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender   string            `json:"sender"`
	Contract string            `json:"contract"`
	CallData map[string]string `json:"call_data"`
	Currency string            `json:"currency"`
}

func (fact *CallContractFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u CallContractFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, u.Sender, u.Contract, u.CallData, u.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op CallContract) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currency.BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CallContract) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
