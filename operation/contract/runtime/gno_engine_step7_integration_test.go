package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const structFieldMapEngineContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Active bool
	Limit int64
}

type User struct {
	Balance int64
	Meta Meta
}

type Config struct {
	Flags map[string]bool
	Users map[string]User
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Flags = map[string]bool{"alpha":true}
	config.Users = map[string]User{
		"alice": User{Balance:1, Meta:Meta{Active:true, Limit:10}},
	}
	return nil
}

func UpdateState(ctx chain.WriteContext, flag string, active bool, owner string, balance int64, userActive bool, limit int64) error {
	if config.Flags == nil {
		config.Flags = map[string]bool{}
	}
	if config.Users == nil {
		config.Users = map[string]User{}
	}
	config.Flags[flag] = active
	config.Users[owner] = User{Balance:balance, Meta:Meta{Active:userActive, Limit:limit}}
	return nil
}
`

func TestGnoWritePathStructWithMapRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(structFieldMapEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0006")
	sender := base.NewStringAddress("sender0006")
	runtimeValue := deriveRuntimeState(contract, structFieldMapEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) {
		return nil, false, nil
	}

	registerSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeRegister,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(120),
			ContractCode: structFieldMapEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertNestedStructBinding(t, registerSnapshot, "config", map[string]interface{}{
		"Flags": map[string]interface{}{
			"alpha": "true",
		},
		"Users": map[string]interface{}{
			"alice": map[string]interface{}{
				"Balance": "1",
				"Meta": map[string]interface{}{
					"Active": "true",
					"Limit":  "10",
				},
			},
		},
	})

	callSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeCall,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(121),
			ContractCode: structFieldMapEngineContractSource,
			Function:     "UpdateState",
			CallData: map[string]string{
				"flag":       "beta",
				"active":     "false",
				"owner":      "bob",
				"balance":    "7",
				"userActive": "false",
				"limit":      "77",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertNestedStructBinding(t, callSnapshot, "config", map[string]interface{}{
		"Flags": map[string]interface{}{
			"alpha": "true",
			"beta":  "false",
		},
		"Users": map[string]interface{}{
			"alice": map[string]interface{}{
				"Balance": "1",
				"Meta": map[string]interface{}{
					"Active": "true",
					"Limit":  "10",
				},
			},
			"bob": map[string]interface{}{
				"Balance": "7",
				"Meta": map[string]interface{}{
					"Active": "false",
					"Limit":  "77",
				},
			},
		},
	})
}
