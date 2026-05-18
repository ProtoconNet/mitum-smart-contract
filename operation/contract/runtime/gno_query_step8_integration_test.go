package runtime

import (
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const sliceQueryContractSource = `package contract
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

func Initialize(ctx chain.ContractContext) error {
	config.Names = []string{"alice"}
	config.Users = []User{User{Meta:Meta{Limit:10}}}
	return nil
}

func AppendState(ctx chain.ContractContext, name string, limit int64) error {
	config.Names = append(config.Names, name)
	config.Users = append(config.Users, User{Meta:Meta{Limit:limit}})
	return nil
}

func GetNameAt(ctx chain.ContractContext, i int) (string, bool) {
	if i < 0 || i >= len(config.Names) {
		return "", false
	}
	return config.Names[i], true
}

func GetUserLimitAt(ctx chain.ContractContext, i int) (int64, bool) {
	if i < 0 || i >= len(config.Users) {
		return 0, false
	}
	return config.Users[i].Meta.Limit, true
}

func GetCount(ctx chain.ContractContext) int64 {
	return int64(len(config.Names))
}
`

func TestGnoQueryPathSliceRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		sliceQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(170),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: sliceQueryContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(171),
				Function:     "AppendState",
				CallData:     map[string]string{"name": "bob", "limit": "77"},
				ContractCode: sliceQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)
	height := states[pstate.SnapshotStateKey(contract)].Height()

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       height,
		ContractCode: sliceQueryContractSource,
		Function:     "GetNameAt",
		CallData:     map[string]string{"i": "1"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetNameAt) returned error: %v", err)
	}
	if got := qr.Result.(string); got != "bob" {
		t.Fatalf("unexpected GetNameAt result: %s", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetNameAt ok=true, got %#v", qr.Ok)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       height,
		ContractCode: sliceQueryContractSource,
		Function:     "GetUserLimitAt",
		CallData:     map[string]string{"i": "1"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetUserLimitAt) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 77 {
		t.Fatalf("unexpected GetUserLimitAt result: %d", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetUserLimitAt ok=true, got %#v", qr.Ok)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       height,
		ContractCode: sliceQueryContractSource,
		Function:     "GetCount",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetCount) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 2 {
		t.Fatalf("unexpected GetCount result: %d", got)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
