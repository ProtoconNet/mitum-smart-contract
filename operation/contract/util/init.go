package util

import (
	"reflect"
)

var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/ProtoconNet/mitum-currency/v3/operation/contract/util/util"] = map[string]reflect.Value{
		// type definitions
		"APICollection":   reflect.ValueOf((*APICollection)(nil)),
		"ContractContext": reflect.ValueOf((*ContractContext)(nil)),
		"GetAccountState": reflect.ValueOf((*GetAccountStateFunc)(nil)),
		"GetCallData":     reflect.ValueOf((*GetCallDataFunc)(nil)),
		"GetDataState":    reflect.ValueOf((*GetDataStateFunc)(nil)),
		"GetSender":       reflect.ValueOf((*GetSenderFunc)(nil)),
		"S":               reflect.ValueOf((*S)(nil)),
		"S1":              reflect.ValueOf((*S1)(nil)),

		// interface wrapper definitions
		"_S": reflect.ValueOf((*_github_com_ProtoconNet_mitum_currency_operation_contract_util_S)(nil)),
	}
}

// _github_com_ProtoconNet_mitum_currency_operation_contract_util_S is an interface wrapper for S type
type _github_com_ProtoconNet_mitum_currency_operation_contract_util_S struct {
	IValue interface{}
	WName  func() string
}

func (W _github_com_ProtoconNet_mitum_currency_operation_contract_util_S) Name() string {
	return W.WName()
}
