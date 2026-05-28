package contract

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	RegisterContractFactHint = hint.MustNewHint("mitum-contract-register-operation-fact-v0.0.1")
	RegisterContractHint     = hint.MustNewHint("mitum-contract-register-operation-v0.0.1")
)

const MaxContractSourceBytes = runtime.MaxTypedContractSourceBytes

type RegisterContractFact struct {
	base.BaseFact
	sender   base.Address
	contract base.Address
	code     string
	callData map[string]string
	currency ctypes.CurrencyID
}

func NewRegisterContractFact(token []byte, sender, contract base.Address, code string, callData map[string]string, currency ctypes.CurrencyID) RegisterContractFact {
	bf := base.NewBaseFact(RegisterContractFactHint, token)
	fact := RegisterContractFact{
		BaseFact: bf,
		sender:   sender,
		contract: contract,
		code:     code,
		callData: callData,
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact RegisterContractFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if len(fact.code) > MaxContractSourceBytes {
		return common.ErrFactInvalid.Wrap(
			common.ErrValueInvalid.Wrap(
				errors.Errorf(
					"contract source exceeds max size: got %d bytes, max %d bytes",
					len(fact.code),
					MaxContractSourceBytes,
				),
			),
		)
	}

	if fact.sender.Equal(fact.contract) {
		return common.ErrFactInvalid.Wrap(common.ErrSelfTarget.Wrap(errors.Errorf("sender %v is same with contract account", fact.sender)))
	}

	if fact.callData == nil {
		return common.ErrFactInvalid.Wrap(
			common.ErrValueInvalid.Wrap(errors.Errorf("callData map not initialized")))
	}
	if err := runtime.ValidateContractCallDataLimits("register init_data", fact.callData); err != nil {
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

func (fact RegisterContractFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact RegisterContractFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact RegisterContractFact) Bytes() []byte {
	var bCallData []byte
	if fact.callData != nil {
		d, _ := json.Marshal(fact.callData)
		bCallData = valuehash.NewSHA256(d).Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		fact.contract.Bytes(),
		[]byte(fact.code),
		bCallData,
		fact.currency.Bytes(),
	)
}

func (fact RegisterContractFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact RegisterContractFact) Sender() base.Address {
	return fact.sender
}

func (fact RegisterContractFact) Contract() base.Address {
	return fact.contract
}

func (fact RegisterContractFact) ContractCode() string {
	return fact.code
}

func (fact RegisterContractFact) Addresses() ([]base.Address, error) {
	return []base.Address{fact.sender, fact.contract}, nil
}

type RegisterContract struct {
	common.BaseOperation
}

func NewRegisterContract(fact RegisterContractFact) (RegisterContract, error) {
	return RegisterContract{
		BaseOperation: common.NewBaseOperation(RegisterContractHint, fact),
	}, nil
}
