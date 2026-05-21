package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const scalarSliceEngineContractSource = `package contract
import "mitum/chain"

var names []string

func Initialize(ctx chain.WriteContext) error {
	names = []string{"alice"}
	return nil
}

func AddName(ctx chain.WriteContext, name string) error {
	names = append(names, name)
	return nil
}
`

const namedStructSliceEngineContractSource = `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users []User

func Initialize(ctx chain.WriteContext) error {
	users = []User{User{Balance:1, Active:true}}
	return nil
}

func AppendUser(ctx chain.WriteContext, balance int64, active bool) error {
	users = append(users, User{Balance:balance, Active:active})
	return nil
}
`

const structFieldSliceEngineContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
}

type User struct {
	Meta Meta
}

type Config struct {
	Names []string
	Users []User
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Names = []string{"alice"}
	config.Users = []User{User{Meta:Meta{Limit:10}}}
	return nil
}

func AppendState(ctx chain.WriteContext, name string, limit int64) error {
	config.Names = append(config.Names, name)
	config.Users = append(config.Users, User{Meta:Meta{Limit:limit}})
	return nil
}
`

func TestGnoWritePathTopLevelScalarSliceRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(scalarSliceEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0008")
	sender := base.NewStringAddress("sender0008")
	runtimeValue := deriveRuntimeState(contract, scalarSliceEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) { return nil, false, nil }

	registerSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(140),
		ContractCode: scalarSliceEngineContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	}, nil)
	assertSliceBinding(t, registerSnapshot, "names", []interface{}{"alice"})

	callSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(141),
		ContractCode: scalarSliceEngineContractSource,
		Function:     "AddName",
		CallData:     map[string]string{"name": "bob"},
	}, mustMarshalSnapshotDoc(t, registerSnapshot))
	assertSliceBinding(t, callSnapshot, "names", []interface{}{"alice", "bob"})
}

func TestGnoWritePathTopLevelNamedStructSliceRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(namedStructSliceEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0009")
	sender := base.NewStringAddress("sender0009")
	runtimeValue := deriveRuntimeState(contract, namedStructSliceEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) { return nil, false, nil }

	registerSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(150),
		ContractCode: namedStructSliceEngineContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	}, nil)
	assertSliceBinding(t, registerSnapshot, "users", []interface{}{
		map[string]interface{}{"Balance": "1", "Active": "true"},
	})

	callSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(151),
		ContractCode: namedStructSliceEngineContractSource,
		Function:     "AppendUser",
		CallData:     map[string]string{"balance": "7", "active": "false"},
	}, mustMarshalSnapshotDoc(t, registerSnapshot))
	assertSliceBinding(t, callSnapshot, "users", []interface{}{
		map[string]interface{}{"Balance": "1", "Active": "true"},
		map[string]interface{}{"Balance": "7", "Active": "false"},
	})
}

func TestGnoWritePathStructFieldSliceRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(structFieldSliceEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0010")
	sender := base.NewStringAddress("sender0010")
	runtimeValue := deriveRuntimeState(contract, structFieldSliceEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) { return nil, false, nil }

	registerSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(160),
		ContractCode: structFieldSliceEngineContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	}, nil)
	assertNestedStructBinding(t, registerSnapshot, "config", map[string]interface{}{
		"Names": []interface{}{"alice"},
		"Users": []interface{}{
			map[string]interface{}{"Meta": map[string]interface{}{"Limit": "10"}},
		},
	})

	callSnapshot := executeGnoWriteForTest(t, encs, getStateFunc, schema, runtimeValue.PackagePath, ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(161),
		ContractCode: structFieldSliceEngineContractSource,
		Function:     "AppendState",
		CallData:     map[string]string{"name": "bob", "limit": "77"},
	}, mustMarshalSnapshotDoc(t, registerSnapshot))
	assertNestedStructBinding(t, callSnapshot, "config", map[string]interface{}{
		"Names": []interface{}{"alice", "bob"},
		"Users": []interface{}{
			map[string]interface{}{"Meta": map[string]interface{}{"Limit": "10"}},
			map[string]interface{}{"Meta": map[string]interface{}{"Limit": "77"}},
		},
	})
}

func assertSliceBinding(t *testing.T, doc SnapshotDoc, name string, expected []interface{}) {
	t.Helper()
	binding := findSnapshotBinding(t, doc, name)
	assertSliceSnapshotValue(t, binding.Value, expected)
}
