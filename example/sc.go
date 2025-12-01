package main

import (
	"fmt"

	"github.com/ProtoconNet/mitum-currency/v3/operation/contract/util"
)

func Initialize(ctx util.ContractContext) (map[string]interface{}, error) {
	sender := ctx.GetSender()
	_, err := ctx.GetAccountState(sender)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func CreateData(ctx util.ContractContext) (map[string]interface{}, error) {
	sender := ctx.GetSender()
	data, _ := ctx.GetDataState(sender)
	if data != nil {
		return nil, fmt.Errorf("already exist data state for data key %v", sender)
	}
	callData := ctx.GetCallData()
	dataArg, found := callData["data"]
	if !found {
		return nil, fmt.Errorf("data not found in call data")
	}
	userData := NewData(
		sender, dataArg, "userAddress",
	)

	return userData, nil
}

func UpdateData(ctx util.ContractContext) (map[string]interface{}, error) {
	sender := ctx.GetSender()
	data, _ := ctx.GetDataState(sender)
	if data == nil {
		return nil, fmt.Errorf("not found data state for data key %v", sender)
	}
	callData := ctx.GetCallData()
	dataArg, found := callData["data"]
	if !found {
		return nil, fmt.Errorf("data not found in call data")
	}
	userData := NewData(
		sender, dataArg, "userAddress",
	)

	return userData, nil
}

type Data map[string]interface{}

func NewData(
	key, value, valueType string,
) Data {
	data := map[string]interface{}{
		"valueType": valueType,
		"key":       key,
		"value":     value,
	}

	return data
}

func (d Data) IsValid([]byte) error {
	return nil
}

func (d Data) Key() *string {
	v, found := d["key"]
	if !found {
		return nil
	}

	str, ok := v.(string)
	if !ok {
		return nil
	}

	return &str
}

func (d Data) Value() *string {
	v, found := d["value"]
	if !found {
		return nil
	}

	str, ok := v.(string)
	if !ok {
		return nil
	}

	return &str
}
