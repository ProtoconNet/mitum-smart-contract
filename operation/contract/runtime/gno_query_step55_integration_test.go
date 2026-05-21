package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const scalarQueryContractSource = `package contract
import "mitum/chain"

var revision int64
var writeHeight int64
var writeReadOnly bool

func Initialize(ctx chain.WriteContext) error {
	revision = 1
	writeHeight = ctx.GetHeight()
	writeReadOnly = ctx.IsReadOnly()
	return nil
}

func Bump(ctx chain.WriteContext, amount int64) error {
	revision = revision + amount
	writeHeight = ctx.GetHeight()
	writeReadOnly = ctx.IsReadOnly()
	return nil
}

func GetRevision(ctx chain.QueryContext) int64 {
	return revision
}

func IsReadOnlyQuery(ctx chain.QueryContext) bool {
	return ctx.IsReadOnly()
}

func GetQueryHeight(ctx chain.QueryContext) int64 {
	return ctx.GetHeight()
}

func GetContractAddress(ctx chain.QueryContext) string {
	return ctx.GetContract()
}

func GetWriteHeight(ctx chain.QueryContext) int64 {
	return writeHeight
}

func GetCurrentHeight(ctx chain.QueryContext) int64 {
	return ctx.GetCurrentHeight()
}

func WasWriteReadOnly(ctx chain.QueryContext) bool {
	return writeReadOnly
}
`

const flatStructQueryContractSource = `package contract
import "mitum/chain"

type Config struct {
	Owner string
	Paused bool
	Limit int64
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Owner = ctx.GetSender()
	config.Paused = false
	config.Limit = 1
	return nil
}

func UpdateConfig(ctx chain.WriteContext, paused bool, limit int64) error {
	config.Paused = paused
	config.Limit = limit
	return nil
}

func GetLimit(ctx chain.QueryContext) int64 {
	return config.Limit
}

func GetOwner(ctx chain.QueryContext) string {
	return config.Owner
}
`

const mapScalarQueryContractSource = `package contract
import "mitum/chain"

var balances map[string]int64

func Initialize(ctx chain.WriteContext) error {
	balances = map[string]int64{"alice":1}
	return nil
}

func AddBalance(ctx chain.WriteContext, owner string, amount int64) error {
	if balances == nil {
		balances = map[string]int64{}
	}
	balances[owner] = balances[owner] + amount
	return nil
}

func GetBalance(ctx chain.QueryContext, owner string) (int64, bool) {
	v, found := balances[owner]
	return v, found
}
`

const mapStructQueryContractSource = `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.WriteContext) error {
	users = map[string]User{
		"alice": User{Balance:1, Active:true},
	}
	return nil
}

func UpdateUser(ctx chain.WriteContext, owner string, balance int64, active bool) error {
	if users == nil {
		users = map[string]User{}
	}
	users[owner] = User{Balance:balance, Active:active}
	return nil
}

func GetUserBalance(ctx chain.QueryContext, owner string) (int64, bool) {
	user, found := users[owner]
	if !found {
		return 0, false
	}
	return user.Balance, true
}

func IsUserActive(ctx chain.QueryContext, owner string) (bool, bool) {
	user, found := users[owner]
	if !found {
		return false, false
	}
	return user.Active, true
}
`

const mutatingQueryContractSource = `package contract
import "mitum/chain"

var revision int64

func Initialize(ctx chain.WriteContext) error {
	revision = 1
	return nil
}

func GetAndBump(ctx chain.QueryContext) int64 {
	revision = revision + 1
	return revision
}
`

func TestGnoQueryPathScalarRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		scalarQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(40),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: scalarQueryContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(41),
				Function:     "Bump",
				CallData:     map[string]string{"amount": "6"},
				ContractCode: scalarQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "GetRevision",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetRevision) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 7 {
		t.Fatalf("unexpected GetRevision result: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "IsReadOnlyQuery",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(IsReadOnlyQuery) returned error: %v", err)
	}
	if got := qr.Result.(bool); !got {
		t.Fatalf("expected read-only query context")
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "GetQueryHeight",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetQueryHeight) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != int64(states[pstate.SnapshotStateKey(contract)].Height()) {
		t.Fatalf("unexpected query context height: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        states[pstate.SnapshotStateKey(contract)].Height(),
		CurrentHeight: base.Height(77),
		ContractCode:  scalarQueryContractSource,
		Function:      "GetCurrentHeight",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetCurrentHeight) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 77 {
		t.Fatalf("unexpected current chain height: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "GetContractAddress",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetContractAddress) returned error: %v", err)
	}
	if got := qr.Result.(string); got != contract.String() {
		t.Fatalf("unexpected query context contract: %q", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "GetWriteHeight",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetWriteHeight) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 41 {
		t.Fatalf("unexpected write context height: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: scalarQueryContractSource,
		Function:     "WasWriteReadOnly",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(WasWriteReadOnly) returned error: %v", err)
	}
	if got := qr.Result.(bool); got {
		t.Fatalf("expected write context to be non-read-only")
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathFlatStructRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		flatStructQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(50),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: flatStructQueryContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(51),
				Function:     "UpdateConfig",
				CallData:     map[string]string{"paused": "true", "limit": "9"},
				ContractCode: flatStructQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: flatStructQueryContractSource,
		Function:     "GetLimit",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetLimit) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 9 {
		t.Fatalf("unexpected GetLimit result: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: flatStructQueryContractSource,
		Function:     "GetOwner",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetOwner) returned error: %v", err)
	}
	if got := qr.Result.(string); got != sender.String() {
		t.Fatalf("unexpected GetOwner result: %s", got)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMapStringScalarRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		mapScalarQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(60),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: mapScalarQueryContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(61),
				Function:     "AddBalance",
				CallData:     map[string]string{"owner": "bob", "amount": "7"},
				ContractCode: mapScalarQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: mapScalarQueryContractSource,
		Function:     "GetBalance",
		CallData:     map[string]string{"owner": "bob"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetBalance) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 7 {
		t.Fatalf("unexpected GetBalance result: %d", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetBalance ok=true, got %#v", qr.Ok)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMapStringStructRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		mapStructQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(70),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: mapStructQueryContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(71),
				Function:     "UpdateUser",
				CallData:     map[string]string{"owner": "bob", "balance": "7", "active": "false"},
				ContractCode: mapStructQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: mapStructQueryContractSource,
		Function:     "GetUserBalance",
		CallData:     map[string]string{"owner": "bob"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetUserBalance) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 7 {
		t.Fatalf("unexpected GetUserBalance result: %d", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetUserBalance ok=true, got %#v", qr.Ok)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: mapStructQueryContractSource,
		Function:     "IsUserActive",
		CallData:     map[string]string{"owner": "bob"},
	})
	if err != nil {
		t.Fatalf("QueryContract(IsUserActive) returned error: %v", err)
	}
	if got := qr.Result.(bool); got {
		t.Fatalf("unexpected IsUserActive result: %v", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected IsUserActive ok=true, got %#v", qr.Ok)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathRejectsMutation(t *testing.T) {
	_, states, contract, sender, _ := prepareQueryTestState(
		t,
		mutatingQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(80),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: mutatingQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: mutatingQueryContractSource,
		Function:     "GetAndBump",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatalf("expected mutating query to fail")
	}
}

func prepareQueryTestState(
	t *testing.T,
	source string,
	requests []ExecuteRequest,
) (ContractSchema, map[string]base.State, base.Address, base.Address, []byte) {
	t.Helper()

	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contractq0001")
	sender := base.NewStringAddress("senderq0001")
	runtimeValue := deriveRuntimeState(contract, source)
	getStateFunc := func(key string) (base.State, bool, error) {
		return nil, false, nil
	}

	var snapshotDoc SnapshotDoc
	var snapshotBytes []byte
	for _, req := range requests {
		req.Contract = contract
		req.Sender = sender
		req.ContractCode = source

		snapshotDoc = executeGnoWriteForTest(
			t,
			encs,
			getStateFunc,
			schema,
			runtimeValue.PackagePath,
			req,
			snapshotBytes,
		)
		snapshotBytes = mustMarshalSnapshotDoc(t, snapshotDoc)
	}

	states := map[string]base.State{
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			requests[0].Height,
			pstate.RuntimeStateKey(contract),
			runtimeValue,
			nil,
			nil,
		),
		pstate.SnapshotStateKey(contract): common.NewBaseState(
			requests[len(requests)-1].Height,
			pstate.SnapshotStateKey(contract),
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, snapshotBytes),
			nil,
			nil,
		),
	}

	return schema, states, contract, sender, append([]byte(nil), snapshotBytes...)
}

func stateGetter(states map[string]base.State) base.GetStateFunc {
	return func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}
}

func assertSnapshotStateUnchanged(t *testing.T, states map[string]base.State, contract base.Address, before []byte) {
	t.Helper()

	snapshotValue, err := pstate.GetSnapshotFromState(states[pstate.SnapshotStateKey(contract)])
	if err != nil {
		t.Fatalf("GetSnapshotFromState returned error: %v", err)
	}

	after := snapshotValue.Snapshot
	if string(before) != string(after) {
		t.Fatalf("expected query to leave snapshot state unchanged\nbefore: %s\nafter:  %s", before, after)
	}
}
