package runtime

import (
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

type AccountReader interface {
	AccountExists(addr string) (bool, error)
}

type ContractReader interface {
	IsContractAccount(addr string) (bool, error)
}

type BalanceReader interface {
	BalanceOf(addr string, currency string) (string, bool, error)
}

type ExecutionContext struct {
	Sender         base.Address
	Contract       base.Address
	Height         base.Height
	CurrentHeight  base.Height
	ReadOnly       bool
	AccountReader  AccountReader
	ContractReader ContractReader
	BalanceReader  BalanceReader
}

func (ctx *ExecutionContext) Validate() error {
	if ctx == nil {
		return errors.Errorf("execution context is nil")
	}
	if ctx.Sender == nil {
		return errors.Errorf("execution context sender is nil")
	}
	if ctx.Contract == nil {
		return errors.Errorf("execution context contract is nil")
	}
	if ctx.CurrentHeight < ctx.Height {
		return errors.Errorf("execution context current height %d is below view height %d", ctx.CurrentHeight, ctx.Height)
	}
	if ctx.AccountReader == nil {
		return errors.Errorf("execution context account reader is nil")
	}
	if ctx.ContractReader == nil {
		return errors.Errorf("execution context contract reader is nil")
	}
	if ctx.BalanceReader == nil {
		return errors.Errorf("execution context balance reader is nil")
	}

	return nil
}

func NewExecutionContext(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	contract base.Address,
	sender base.Address,
	height base.Height,
	readOnly bool,
) (*ExecutionContext, error) {
	return NewExecutionContextWithCurrentHeight(
		encs,
		getStateFunc,
		contract,
		sender,
		height,
		height,
		readOnly,
	)
}

func NewExecutionContextWithCurrentHeight(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	contract base.Address,
	sender base.Address,
	height base.Height,
	currentHeight base.Height,
	readOnly bool,
) (*ExecutionContext, error) {
	ctx := &ExecutionContext{
		Sender:         sender,
		Contract:       contract,
		Height:         height,
		CurrentHeight:  currentHeight,
		ReadOnly:       readOnly,
		AccountReader:  NewStateAccountReader(encs, getStateFunc),
		ContractReader: NewStateContractReader(encs, getStateFunc),
		BalanceReader:  NewStateBalanceReader(encs, getStateFunc),
	}

	if err := ctx.Validate(); err != nil {
		return nil, err
	}

	return ctx, nil
}

type StateAccountReader struct {
	encs         encoder.Encoders
	getStateFunc base.GetStateFunc
}

func NewStateAccountReader(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
) StateAccountReader {
	return StateAccountReader{
		encs:         encs,
		getStateFunc: getStateFunc,
	}
}

func (r StateAccountReader) AccountExists(addr string) (bool, error) {
	return GetAccountStateFunc(addr, r.encs, r.getStateFunc)
}

type StateContractReader struct {
	encs         encoder.Encoders
	getStateFunc base.GetStateFunc
}

func NewStateContractReader(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
) StateContractReader {
	return StateContractReader{
		encs:         encs,
		getStateFunc: getStateFunc,
	}
}

func (r StateContractReader) IsContractAccount(addr string) (bool, error) {
	address, err := base.DecodeAddress(addr, r.encs.JSON())
	if err != nil {
		return false, errors.Errorf("failed to decode address, %v", addr)
	}

	st, found, err := r.getStateFunc(cestate.StateKeyContractAccount(address))
	switch {
	case err != nil:
		return false, errors.Errorf("contract account, %v: %v", addr, err)
	case !found:
		return false, nil
	default:
		if _, err := cestate.StateContractAccountValue(st); err != nil {
			return false, errors.Errorf("contract account, %v: %v", addr, err)
		}
	}

	return true, nil
}
