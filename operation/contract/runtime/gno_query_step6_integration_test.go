package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const nestedQueryContractSource = `package contract
import "mitum/chain"

type Limits struct {
	Daily int64
	Max int64
}

type Config struct {
	Owner string
	Limits Limits
}

type Meta struct {
	Active bool
	Limit int64
}

type User struct {
	Balance int64
	Meta Meta
}

var config Config
var users map[string]User

func Initialize(ctx chain.WriteContext) error {
	config.Owner = ctx.GetSender()
	config.Limits.Daily = 1
	config.Limits.Max = 2
	users = map[string]User{
		"alice": User{Balance:1, Meta:Meta{Active:true, Limit:10}},
	}
	return nil
}

func UpdateState(ctx chain.WriteContext, daily int64, max int64, owner string, balance int64, active bool, limit int64) error {
	config.Limits.Daily = daily
	config.Limits.Max = max
	if users == nil {
		users = map[string]User{}
	}
	users[owner] = User{Balance:balance, Meta:Meta{Active:active, Limit:limit}}
	return nil
}

func GetDailyLimit(ctx chain.QueryContext) int64 {
	return config.Limits.Daily
}

func GetUserMetaLimit(ctx chain.QueryContext, owner string) (int64, bool) {
	user, found := users[owner]
	if !found {
		return 0, false
	}
	return user.Meta.Limit, true
}
`

func TestGnoQueryPathNestedStructRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		nestedQueryContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(110),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: nestedQueryContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(111),
				Function: "UpdateState",
				CallData: map[string]string{
					"daily":   "7",
					"max":     "11",
					"owner":   "bob",
					"balance": "9",
					"active":  "false",
					"limit":   "77",
				},
				ContractCode: nestedQueryContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: nestedQueryContractSource,
		Function:     "GetDailyLimit",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetDailyLimit) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 7 {
		t.Fatalf("unexpected GetDailyLimit result: %d", got)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: nestedQueryContractSource,
		Function:     "GetUserMetaLimit",
		CallData:     map[string]string{"owner": "bob"},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetUserMetaLimit) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 77 {
		t.Fatalf("unexpected GetUserMetaLimit result: %d", got)
	}
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetUserMetaLimit ok=true, got %#v", qr.Ok)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
