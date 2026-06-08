package contract

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	CallContractFactHint = hint.MustNewHint("mitum-contract-call-operation-fact-v0.0.1")
	CallContractHint     = hint.MustNewHint("mitum-contract-call-operation-v0.0.1")
)

const (
	MaxCallContractItems           = runtime.MaxContractCallItems
	MaxCallContractItemsTotalBytes = runtime.MaxContractCallItemsTotalBytes
)

type CallContractItem struct {
	function string
	callData map[string]string
}

func NewCallContractItem(function string, callData map[string]string) CallContractItem {
	return CallContractItem{
		function: function,
		callData: copyStringMap(callData),
	}
}

func (it CallContractItem) Bytes() []byte {
	d, _ := json.Marshal(normalizeStringMap(it.callData))

	return util.ConcatBytesSlice(
		valuehash.NewSHA256([]byte(it.function)).Bytes(),
		valuehash.NewSHA256(d).Bytes(),
	)
}

func (it CallContractItem) IsValid([]byte) error {
	if it.function == "" {
		return common.ErrItemInvalid.Wrap(common.ErrValueInvalid.Wrap(errors.Errorf("function is empty")))
	}
	if _, found := it.callData["function"]; found {
		return common.ErrItemInvalid.Wrap(common.ErrValueInvalid.Wrap(
			errors.Errorf("callData must not include function selector key")))
	}
	if err := runtime.ValidateContractCallDataLimits("call item call_data", it.callData); err != nil {
		return common.ErrItemInvalid.Wrap(common.ErrValueInvalid.Wrap(err))
	}

	return nil
}

func (it CallContractItem) Function() string {
	return it.function
}

func (it CallContractItem) CallData() map[string]string {
	return copyStringMap(it.callData)
}

func (it CallContractItem) Rebuild() CallContractItem {
	it.callData = copyStringMap(it.callData)

	return it
}

type CallContractFact struct {
	base.BaseFact
	sender   base.Address
	contract base.Address
	items    []CallContractItem
	currency types.CurrencyID
}

func NewCallContractFact(
	token []byte, sender, contract base.Address, callData map[string]string, currency types.CurrencyID,
) CallContractFact {
	items := normalizeLegacyCallData(callData)

	return newCallContractFact(token, sender, contract, items, currency)
}

func NewCallContractFactWithItems(
	token []byte, sender, contract base.Address, items []CallContractItem, currency types.CurrencyID,
) CallContractFact {
	return newCallContractFact(token, sender, contract, items, currency)
}

func newCallContractFact(
	token []byte,
	sender, contract base.Address,
	items []CallContractItem,
	currency types.CurrencyID,
) CallContractFact {
	bf := base.NewBaseFact(CallContractFactHint, token)
	fact := CallContractFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		items:    copyCallContractItems(items),
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CallContractFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(
			common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if len(fact.items) < 1 {
		return common.ErrFactInvalid.Wrap(
			common.ErrArrayLen.Wrap(errors.Errorf("empty items")))
	}
	if len(fact.items) > MaxCallContractItems {
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(
			errors.Errorf("items, %d over max, %d", len(fact.items), MaxCallContractItems)))
	}
	if len(fact.items) == 1 {
		if err := runtime.ValidateContractCallDataLimits("call_data", fact.items[0].callData); err != nil {
			return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(err))
		}
		if err := runtime.ValidateContractCallDataLimits("call_data", fact.items[0].rebuildLegacyCallData()); err != nil {
			return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(err))
		}
	}
	for i := range fact.items {
		if err := util.CheckIsValiders(nil, false, fact.items[i]); err != nil {
			return common.ErrFactInvalid.Wrap(err)
		}
	}
	if err := runtime.ValidateContractCallItemsLimits("call items", fact.runtimeCallItems()); err != nil {
		return common.ErrFactInvalid.Wrap(common.ErrValueInvalid.Wrap(err))
	}

	if err := util.CheckIsValiders(nil, false,
		fact.sender,
		fact.contract,
		fact.currency,
	); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact CallContractFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CallContractFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CallContractFact) Bytes() []byte {
	var bItems []byte
	switch len(fact.items) {
	case 0:
	case 1:
		d, _ := json.Marshal(fact.items[0].rebuildLegacyCallData())
		bItems = valuehash.NewSHA256(d).Bytes()
	default:
		is := make([][]byte, len(fact.items))
		for i := range fact.items {
			is[i] = fact.items[i].Bytes()
		}
		bItems = valuehash.NewSHA256(util.ConcatBytesSlice(is...)).Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		bItems,
		fact.currency.Bytes(),
	)
}

func (fact CallContractFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CallContractFact) Sender() base.Address {
	return fact.sender
}

func (fact CallContractFact) Contract() base.Address {
	return fact.contract
}

func (fact CallContractFact) Items() []CallContractItem {
	return copyCallContractItems(fact.items)
}

func (fact CallContractFact) CallData() map[string]string {
	if len(fact.items) != 1 {
		return nil
	}

	return fact.items[0].rebuildLegacyCallData()
}

func (fact CallContractFact) Addresses() ([]base.Address, error) {
	return []base.Address{fact.sender, fact.contract}, nil
}

type CallContract struct {
	common.BaseOperation
}

func NewCallContract(fact CallContractFact) (CallContract, error) {
	return CallContract{
		BaseOperation: common.NewBaseOperation(CallContractHint, fact),
	}, nil
}

func (fact CallContractFact) runtimeCallItems() []runtime.ExecuteCallItem {
	items := make([]runtime.ExecuteCallItem, len(fact.items))
	for i := range fact.items {
		items[i] = runtime.ExecuteCallItem{
			Function: fact.items[i].function,
			CallData: copyStringMap(fact.items[i].callData),
		}
	}

	return items
}

func (it CallContractItem) rebuildLegacyCallData() map[string]string {
	callData := copyStringMap(it.callData)
	callData["function"] = it.function

	return callData
}

func normalizeLegacyCallData(callData map[string]string) []CallContractItem {
	raw := copyStringMap(callData)
	itemCallData := make(map[string]string, len(raw))

	function := raw["function"]
	for key, value := range raw {
		if key == "function" {
			continue
		}
		itemCallData[key] = value
	}

	return []CallContractItem{NewCallContractItem(function, itemCallData)}
}

func copyCallContractItems(items []CallContractItem) []CallContractItem {
	if items == nil {
		return nil
	}

	out := make([]CallContractItem, len(items))
	for i := range items {
		out[i] = items[i].Rebuild()
	}

	return out
}

func normalizeStringMap(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}

	return m
}

func copyStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for key, value := range m {
		out[key] = value
	}

	return out
}
