package contract

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	sdk "github.com/ProtoconNet/mitum-currency/v3/operation/contract/util"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var registerContractProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterContractProcessor)
	},
}

var initializeFuncName = "Initialize"

type RegisterContractProcessor struct {
	*base.BaseOperationProcessor
	encs *encoder.Encoders
}

func NewRegisterContractProcessor(encs encoder.Encoders) currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new RegisterContractProcessor")

		nopp := registerContractProcessorPool.Get()
		opp, ok := nopp.(*RegisterContractProcessor)
		if !ok {
			return nil, errors.Errorf("expected RegisterContractProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		if opp.encs == nil {
			opp.encs = &encs
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *RegisterContractProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(RegisterContractFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", RegisterContractFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	_, cSt, aErr, cErr := cstate.ExistsCAccount(fact.Contract(), "contract", true, true, getStateFunc)
	if aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", cErr)), nil
	}

	ca, err := cestate.CheckCAAuthFromState(cSt, fact.Sender())
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if ca == nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMValueInvalid).Errorf(
				"contract account value is nil")), nil
	}

	if ca.IsActive() {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf(
				"contract account %v has already been activated", fact.Contract())), nil
	}

	if found, _ := cstate.CheckNotExistsState(pstate.DesignStateKey(fact.Contract()), getStateFunc); found {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf("wasm service for contract account %v",
				fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *RegisterContractProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(RegisterContractFact)

	if err := ValidateContract(fact.ContractCode()); err != nil {
		return nil, err, nil
	}

	var sts []base.StateMergeValue

	results, bErr := ExecuteContract(
		*opp.encs, getStateFunc, fact.Contract(), fact.Sender(), fact.ContractCode(), initializeFuncName, fact.callData,
	)
	if bErr != nil {
		return nil, bErr, nil
	}

	var result map[string]interface{}
	var ok bool
	if !results[0].IsNil() {
		result, ok = results[0].Interface().(map[string]interface{})
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"initialize must return map[string]interface{}, but got %T; %v", results[0].Interface()), nil
		}
	}

	var err error
	if !results[1].IsNil() {
		err, ok = results[1].Interface().(error)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"initialize did not return an error as expected, got %T", results[1].Interface()), nil
		}
	}

	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to initialize contract code at %v; %v", fact.Contract(), err), nil
	}
	if result != nil {
		_, found := result["valueType"]
		if !found {
			return nil, base.NewBaseOperationProcessReasonError(
				"valueType not found from Initialize result of contract code at %v", fact.Contract()), nil
		}
		key, found := result["key"]
		if !found {
			return nil, base.NewBaseOperationProcessReasonError(
				"key not found from Initialize result of contract code at %v", fact.Contract()), nil
		}
		stKey, ok := key.(string)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"key type expected string, but %T", key), nil
		}
		sts = append(sts, cstate.NewStateMergeValue(
			pstate.DataStateKey(fact.Contract(), stKey),
			pstate.NewDataStateValue(result),
		))
	}

	design := ptypes.NewDesign(fact.ContractCode())
	if err := design.IsValid(nil); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("invalid contract design, %q; %w", fact.Contract(), err), nil
	}

	sts = append(sts, cstate.NewStateMergeValue(
		pstate.DesignStateKey(fact.Contract()),
		pstate.NewDesignStateValue(design),
	))

	st, _ := cstate.ExistsState(cestate.StateKeyContractAccount(fact.Contract()), "contract account", getStateFunc)
	ca, _ := cestate.StateContractAccountValue(st)
	nca := ca.SetIsActive(true)

	sts = append(sts, cstate.NewStateMergeValue(
		cestate.StateKeyContractAccount(fact.Contract()),
		cestate.NewContractAccountStateValue(nca),
	))

	return sts, nil, nil
}

func (opp *RegisterContractProcessor) Close() error {
	registerContractProcessorPool.Put(opp)

	return nil
}

func NewGetAccountStateFunc(encs encoder.Encoders, getStateFunc base.GetStateFunc) sdk.GetAccountStateFunc {
	return func(addr string) (bool, error) {
		return GetAccountStateFunc(addr, encs, getStateFunc)
	}
}

func GetAccountStateFunc(addr string, encs encoder.Encoders, getStateFunc base.GetStateFunc) (bool, error) {
	address, err := base.DecodeAddress(addr, encs.JSON())
	if err != nil {
		return false, errors.Errorf("failed to decode address, %v", addr)
	}

	var st base.State
	var found bool
	k := currency.AccountStateKey(address)
	switch st, found, err = getStateFunc(k); {
	case err != nil:
		return false, errors.Errorf("account, %v: %v", addr, err)
	case !found:
		return false, errors.Errorf("account, %v", addr)
	default:
		_, err = currency.LoadAccountStateValue(st)
		if err != nil {
			return false, errors.Errorf("account, %v: %v", addr, err)
		}
	}
	return true, nil
}

func NewGetDataStateFunc(contract base.Address, encs encoder.Encoders, getStateFunc base.GetStateFunc) sdk.GetDataStateFunc {
	return func(key string) (map[string]interface{}, error) {
		return GetDataStateFunc(contract, key, encs, getStateFunc)
	}
}

func GetDataStateFunc(
	contract base.Address, dataKey string, encs encoder.Encoders, getStateFunc base.GetStateFunc,
) (map[string]interface{}, error) {
	k := pstate.DataStateKey(contract, dataKey)
	var data map[string]interface{}
	switch st, found, err := getStateFunc(k); {
	case err != nil:
		return nil, errors.Errorf("data key, %v: %v", dataKey, err)
	case !found:
		return nil, errors.Errorf("data key, %v", dataKey)
	default:
		data, err = pstate.GetDataFromState(st)
		if err != nil {
			return nil, errors.Errorf("data key, %v: %v", dataKey, err)
		}
	}
	return data, nil
}

func ValidateContract(sourceCode string) base.OperationProcessReasonError {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", sourceCode, 0)
	if err != nil {
		return base.NewBaseOperationProcessReasonError(
			"failed to parse contract code: %w", err)
	}

	var validationErr base.OperationProcessReasonError

	ast.Inspect(node, func(n ast.Node) bool {
		if validationErr != nil {
			return false
		}

		switch n.(type) {
		case *ast.GoStmt:
			validationErr = base.NewBaseOperationProcessReasonError(
				"failed to validate contract code: 'go' routines are not allowed in this contract")
			return false

		case *ast.ForStmt:
			validationErr = base.NewBaseOperationProcessReasonError(
				"failed to validate contract code: 'for' loops are not allowed in this contract")
			return false

		case *ast.RangeStmt:
			validationErr = base.NewBaseOperationProcessReasonError(
				"failed to validate contract code: 'range' loops are not allowed in this contract")
			return false
		}

		//case *ast.MapType:
		//	validationErr = base.NewBaseOperationProcessReasonError(
		//		"failed to validate contract code: 'map' types are not allowed in this contract")
		//	return false
		//}

		return true
	})

	return validationErr
}

func ExecuteContract(
	encs encoder.Encoders, getStateFunc base.GetStateFunc, contract, sender base.Address,
	contractCode, method string, callData map[string]string,
) (results []reflect.Value, berr base.OperationProcessReasonError) {
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				berr = base.NewBaseOperationProcessReasonError(
					"recovered: %v", r)
			}
			close(done)
		}()

		i := interp.New(interp.Options{
			Env: os.Environ(),
		})

		err := i.Use(stdlib.Symbols)
		if err != nil {
			berr = base.NewBaseOperationProcessReasonError(
				"failed to use stdlib Symbols: %w", contract, err)
			return
		}
		err = i.Use(sdk.Symbols)
		if err != nil {
			berr = base.NewBaseOperationProcessReasonError(
				"failed to use sdk Symbols: %w", contract, err)
			return
		}

		_, err = i.Eval(contractCode)
		if err != nil {
			berr = base.NewBaseOperationProcessReasonError(
				"failed to load contract code of %v: %w", contract, err)
			return
		}

		fname := fmt.Sprintf("main.%v", method)
		FuncInitialize, err := i.Eval(fname)
		if err != nil {
			berr = base.NewBaseOperationProcessReasonError(
				"failed to lookup %v function of %v: %w", fname, contract, err)
			return
		}

		ctx := NewContractContext(encs, getStateFunc, contract, sender, callData)
		args := []reflect.Value{reflect.ValueOf(ctx)}
		results = FuncInitialize.Call(args)
		if len(results) != 2 {
			berr = base.NewBaseOperationProcessReasonError(
				"expected 2 results, got %d", len(results))
			return
		}
		if results[1].Interface() != nil {
			var ok bool
			err, ok = results[1].Interface().(error)
			if !ok {
				berr = base.NewBaseOperationProcessReasonError(
					"failed to fetch error of %v function of %v: %w", fname, contract, err)
				return
			}
			berr = base.NewBaseOperationProcessReasonError(
				"failed to execute of %v function of %v: %w", fname, contract, err)
			return
		}
	}()

	select {
	case <-done:
		if berr != nil {
			results = nil
			return
		}
	case <-time.After(2 * time.Second):
		berr = base.NewBaseOperationProcessReasonError(
			"execute function time out")
		results = nil
		return
	}

	return
}

func NewContractContext(
	encs encoder.Encoders, getStateFunc base.GetStateFunc, contract, sender base.Address, callData map[string]string,
) sdk.ContractContext {
	if callData == nil {
		callData = make(map[string]string)
	}
	return sdk.ContractContext{
		GetAccountState: NewGetAccountStateFunc(encs, getStateFunc),
		GetDataState:    NewGetDataStateFunc(contract, encs, getStateFunc),
		GetSender:       func() string { return sender.String() },
		GetCallData:     func() map[string]string { return callData },
	}
}
