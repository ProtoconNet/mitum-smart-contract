package contract

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type RegisterContractFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender   base.Address      `json:"sender"`
	Contract base.Address      `json:"contract"`
	Code     string            `json:"code"`
	CallData map[string]string `json:"call_data"`
	Currency types.CurrencyID  `json:"currency"`
}

func (fact RegisterContractFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RegisterContractFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		Code:                  fact.code,
		CallData:              fact.callData,
		Currency:              fact.currency,
	})
}

type RegisterContractFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender   string            `json:"sender"`
	Contract string            `json:"contract"`
	Code     string            `json:"code"`
	CallData map[string]string `json:"call_data"`
	Currency string            `json:"currency"`
}

func (fact *RegisterContractFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u RegisterContractFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, u.Sender, u.Contract, u.Code, u.CallData, u.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op RegisterContract) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currency.BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *RegisterContract) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
