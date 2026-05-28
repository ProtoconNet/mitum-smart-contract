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

type CallContractFact struct {
	base.BaseFact
	sender   base.Address
	contract base.Address
	callData map[string]string
	currency types.CurrencyID
}

func NewCallContractFact(
	token []byte, sender, contract base.Address, callData map[string]string, currency types.CurrencyID,
) CallContractFact {
	bf := base.NewBaseFact(CallContractFactHint, token)
	fact := CallContractFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		callData: callData,
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

	if fact.callData == nil {
		return common.ErrFactInvalid.Wrap(
			common.ErrValueInvalid.Wrap(errors.Errorf("callData map not initialized")))
	}
	if err := runtime.ValidateContractCallDataLimits("call_data", fact.callData); err != nil {
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
	var bCallData []byte
	if fact.callData != nil {
		d, _ := json.Marshal(fact.callData)
		bCallData = valuehash.NewSHA256(d).Bytes()
	}
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		bCallData,
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
