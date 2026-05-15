package runtime

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path"
	"reflect"
	"strconv"
	"strings"

	sdk "github.com/ProtoconNet/mitum-currency/v3/operation/contract/util"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const contractSDKImport = "github.com/ProtoconNet/mitum-currency/v3/operation/contract/util"

var allowedContractStdlibImports = map[string]struct{}{
	"fmt":           {},
	"errors":        {},
	"strings":       {},
	"strconv":       {},
	"bytes":         {},
	"encoding/json": {},
	"encoding/hex":  {},
}

type yaegiEngine struct{}

func NewYaegiEngine() ContractEngine {
	return yaegiEngine{}
}

func (yaegiEngine) ValidateContract(sourceCode string) base.OperationProcessReasonError {
	return validateContract(sourceCode)
}

func (yaegiEngine) ExecuteContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
) (ExecuteResult, base.OperationProcessReasonError) {
	results, berr := executeRawYaegi(encs, getStateFunc, req)
	if berr != nil {
		return ExecuteResult{}, berr
	}

	return parseYaegiResults(req.Contract, req.Function, results)
}

func (yaegiEngine) QueryContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req QueryRequest,
) (QueryResult, base.OperationProcessReasonError) {
	return QueryResult{}, base.NewBaseOperationProcessReasonError(
		"legacy yaegi contract query is not supported",
	)
}

func parseYaegiResults(
	contract base.Address,
	function string,
	results []reflect.Value,
) (ExecuteResult, base.OperationProcessReasonError) {
	if len(results) != 2 {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
			"expected 2 results, got %d", len(results),
		)
	}

	var result map[string]interface{}
	var ok bool

	if !results[0].IsNil() {
		result, ok = results[0].Interface().(map[string]interface{})
		if !ok {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
				"%v function must return map[string]interface{}, but got %T",
				function,
				results[0].Interface(),
			)
		}
	}

	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
				"%v function did not return an error as expected, got %T",
				function,
				results[1].Interface(),
			)
		}

		return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
			"failed to execute %v function of %v: %v",
			function,
			contract,
			err,
		)
	}

	if result != nil {
		if err := validateContractResultData(result); err != nil {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
				"invalid %v result of contract code at %v: %v",
				function,
				contract,
				err,
			)
		}
	}

	return ExecuteResult{
		Engine: pstate.RuntimeEngineYaegi,
		Data:   result,
	}, nil
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

func validateContractResultData(data map[string]interface{}) error {
	if data == nil {
		return nil
	}

	if _, err := json.Marshal(data); err != nil {
		return errors.Wrap(err, "contract result is not JSON-serializable")
	}

	return nil
}

func isAllowedContractImport(importPath string) bool {
	if importPath == contractSDKImport {
		return true
	}

	_, ok := allowedContractStdlibImports[importPath]
	return ok
}

func validateContract(sourceCode string) base.OperationProcessReasonError {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", sourceCode, 0)
	if err != nil {
		return base.NewBaseOperationProcessReasonError(
			"failed to parse contract code: %w", err)
	}

	for _, imp := range node.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return base.NewBaseOperationProcessReasonError(
				"failed to validate contract import: %w", err)
		}

		if strings.HasPrefix(p, ".") {
			return base.NewBaseOperationProcessReasonError(
				"relative imports are not allowed: %s", p)
		}

		if !isAllowedContractImport(p) {
			return base.NewBaseOperationProcessReasonError(
				"import not allowed in contract: %s", p)
		}
	}

	var validationErr base.OperationProcessReasonError

	ast.Inspect(node, func(n ast.Node) bool {
		if validationErr != nil {
			return false
		}

		switch v := n.(type) {
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

		case *ast.FuncDecl:
			if v.Name != nil && v.Name.Name == "init" {
				validationErr = base.NewBaseOperationProcessReasonError(
					"failed to validate contract code: 'init()' is not allowed in this contract")
				return false
			}

		case *ast.CallExpr:
			if ident, ok := v.Fun.(*ast.Ident); ok && ident.Name == "recover" {
				validationErr = base.NewBaseOperationProcessReasonError(
					"failed to validate contract code: 'recover()' is not allowed in this contract")
				return false
			}
		}

		return true
	})

	if validationErr != nil {
		return validationErr
	}

	schema, err := AnalyzeContractSchema(sourceCode)
	if err != nil {
		return base.NewBaseOperationProcessReasonError(
			"failed to analyze contract schema: %v", err,
		)
	}

	if schema.Mode == SchemaModeTypedArgs {
		return base.NewBaseOperationProcessReasonError(
			"typed-args contract schema detected, but current yaegi runtime still supports only legacy map ABI",
		)
	}

	return nil
}

func filterExportsByImportPath(values interp.Exports, allowed map[string]struct{}) interp.Exports {
	filtered := interp.Exports{}

	for k, v := range values {
		if k == "." {
			filtered[k] = v
			continue
		}

		importPath := path.Dir(k)
		if _, ok := allowed[importPath]; ok {
			filtered[k] = v
		}
	}

	return filtered
}

func executeRawYaegi(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
) (results []reflect.Value, berr base.OperationProcessReasonError) {
	defer func() {
		if r := recover(); r != nil {
			berr = base.NewBaseOperationProcessReasonError(
				"recovered: %v", r)
			results = nil
		}
	}()

	i := interp.New(interp.Options{
		GoPath: build.Default.GOPATH,
	})

	err := i.Use(filterExportsByImportPath(stdlib.Symbols, allowedContractStdlibImports))
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to use filtered stdlib Symbols: %w", err)
	}

	err = i.Use(sdk.Symbols)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to use sdk Symbols: %w", err)
	}

	_, err = i.Eval(req.ContractCode)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to load contract code of %v: %w", req.Contract, err)
	}

	fname := fmt.Sprintf("main.%v", req.Function)
	FuncInitialize, err := i.Eval(fname)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to lookup %v function of %v: %w", fname, req.Contract, err)
	}

	ctx := newContractContext(encs, getStateFunc, req.Contract, req.Sender, req.CallData)
	args := []reflect.Value{reflect.ValueOf(ctx)}
	results = FuncInitialize.Call(args)

	if len(results) != 2 {
		return nil, base.NewBaseOperationProcessReasonError(
			"expected 2 results, got %d", len(results))
	}

	if results[1].Interface() != nil {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"failed to fetch error of %v function of %v: %v", fname, req.Contract, results[1].Interface())
		}

		return nil, base.NewBaseOperationProcessReasonError(
			"failed to execute of %v function of %v: %v", fname, req.Contract, err)
	}

	return results, nil
}

func newContractContext(
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
