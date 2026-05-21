package runtime

import (
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const structFieldMapQueryContractSource = `package contract
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

func GetFlag(ctx chain.QueryContext, name string) (bool, bool) {
	v, found := config.Flags[name]
	return v, found
}

func GetUserLimit(ctx chain.QueryContext, name string) (int64, bool) {
	user, found := config.Users[name]
	if !found {
		return 0, false
	}
	return user.Meta.Limit, true
}
`

func TestGnoQueryPathStructWithMapRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		structFieldMapQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(130),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: structFieldMapQueryContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(131),
				Function: "UpdateState",
				CallData: map[string]string{
					"flag":       "beta",
					"active":     "false",
					"owner":      "bob",
					"balance":    "7",
					"userActive": "false",
					"limit":      "77",
				},
				ContractCode: structFieldMapQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: structFieldMapQueryContractSource,
		Function:     "GetFlag",
		CallData:     map[string]string{"name": "beta"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetFlag) returned error: %v", err)
	}
	if got := qr.Result.(bool); got {
		t.Fatalf("unexpected GetFlag result: %v", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetFlag ok=true, got %#v", qr.Ok)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: structFieldMapQueryContractSource,
		Function:     "GetUserLimit",
		CallData:     map[string]string{"name": "bob"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetUserLimit) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 77 {
		t.Fatalf("unexpected GetUserLimit result: %d", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetUserLimit ok=true, got %#v", qr.Ok)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
