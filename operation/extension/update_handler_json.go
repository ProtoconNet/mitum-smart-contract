package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type UpdateHandlerFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender   base.Address     `json:"sender"`
	Contract base.Address     `json:"contract"`
	Handlers []base.Address   `json:"handlers"`
	Currency types.CurrencyID `json:"currency"`
}

func (fact UpdateHandlerFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateHandlerFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		Handlers:              fact.handlers,
		Currency:              fact.currency,
	})
}

type UpdatHandlerFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender   string   `json:"sender"`
	Contract string   `json:"contract"`
	Handlers []string `json:"handlers"`
	Currency string   `json:"currency"`
}

func (fact *UpdateHandlerFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf UpdatHandlerFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, uf.Sender, uf.Contract, uf.Handlers, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op UpdateHandler) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currency.BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateHandler) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
