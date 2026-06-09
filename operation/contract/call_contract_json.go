package contract

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CallContractFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender   base.Address       `json:"sender"`
	Contract base.Address       `json:"contract"`
	CallData map[string]string  `json:"call_data,omitempty"`
	Items    []CallContractItem `json:"items,omitempty"`
	Currency types.CurrencyID   `json:"currency"`
}

func (fact CallContractFact) MarshalJSON() ([]byte, error) {
	var callData map[string]string
	var items []CallContractItem
	if len(fact.items) == 1 {
		callData = fact.items[0].rebuildLegacyCallData()
	} else {
		items = fact.items
	}

	return util.MarshalJSON(CallContractFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Contract:              fact.contract,
		CallData:              callData,
		Items:                 items,
		Currency:              fact.currency,
	})
}

type CallContractFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender   string           `json:"sender"`
	Contract string           `json:"contract"`
	CallData *json.RawMessage `json:"call_data"`
	Items    *json.RawMessage `json:"items"`
	Currency string           `json:"currency"`
}

func (fact *CallContractFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u CallContractFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	switch {
	case u.CallData != nil && u.Items != nil:
		return common.DecorateError(
			common.ErrValueInvalid.Errorf("call_data and items cannot both be set"),
			common.ErrDecodeJson,
			*fact,
		)
	case u.Items != nil:
		var items []CallContractItem
		if err := enc.Unmarshal(*u.Items, &items); err != nil {
			return common.DecorateError(err, common.ErrDecodeJson, *fact)
		}
		if err := fact.unpackItems(enc, u.Sender, u.Contract, items, u.Currency); err != nil {
			return common.DecorateError(err, common.ErrDecodeJson, *fact)
		}
	default:
		var callData map[string]string
		if u.CallData != nil {
			if err := enc.Unmarshal(*u.CallData, &callData); err != nil {
				return common.DecorateError(err, common.ErrDecodeJson, *fact)
			}
		}
		if err := fact.unpackLegacy(enc, u.Sender, u.Contract, callData, u.Currency); err != nil {
			return common.DecorateError(err, common.ErrDecodeJson, *fact)
		}
	}

	return nil
}

type CallContractItemJSONMarshaler struct {
	hint.BaseHinter
	Function string            `json:"function"`
	CallData map[string]string `json:"call_data"`
}

func (it CallContractItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CallContractItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Function:   it.function,
		CallData:   normalizeStringMap(it.callData),
	})
}

type CallContractItemJSONUnmarshaler struct {
	Hint     string            `json:"_hint"`
	Function string            `json:"function"`
	CallData map[string]string `json:"call_data"`
}

func (it *CallContractItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u CallContractItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	n, err := decodeCallContractItemHint(u.Hint, u.Function, u.CallData)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}
	*it = n

	return nil
}

func (it *CallContractItem) UnmarshalJSON(b []byte) error {
	var u CallContractItemJSONUnmarshaler
	if err := json.Unmarshal(b, &u); err != nil {
		return err
	}

	n, err := decodeCallContractItemHint(u.Hint, u.Function, u.CallData)
	if err != nil {
		return err
	}
	*it = n

	return nil
}

func decodeCallContractItemHint(s, function string, callData map[string]string) (CallContractItem, error) {
	if s == "" {
		return NewCallContractItem(function, callData), nil
	}

	ht, err := hint.ParseHint(s)
	if err != nil {
		return CallContractItem{}, err
	}

	return newCallContractItem(ht, function, callData), nil
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
