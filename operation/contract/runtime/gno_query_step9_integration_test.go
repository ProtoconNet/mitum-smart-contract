package runtime

import (
	"reflect"
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const complexQueryResultContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
	Flags map[string]bool
	Aliases []string
}

type User struct {
	Balance int64
	Meta Meta
}

type Config struct {
	Owner string
	FeatureFlags map[string]bool
	Users map[string]User
	Aliases []string
	Watchers []User
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Owner = ctx.GetSender()
	config.FeatureFlags = map[string]bool{"alpha":true}
	config.Users = map[string]User{
		"alice": User{Balance:1, Meta:Meta{Limit:10, Flags:map[string]bool{"vip":true}, Aliases:[]string{"a1"}}},
	}
	config.Aliases = []string{"root"}
	config.Watchers = []User{
		User{Balance:2, Meta:Meta{Limit:20, Flags:map[string]bool{"watch":true}, Aliases:[]string{"w1"}}},
	}
	return nil
}

func UpsertUser(ctx chain.WriteContext, name string, balance int64, limit int64, beta bool, alias string, watcherAlias string) error {
	if config.FeatureFlags == nil {
		config.FeatureFlags = map[string]bool{}
	}
	if config.Users == nil {
		config.Users = map[string]User{}
	}
	config.FeatureFlags["beta"] = beta
	config.Users[name] = User{Balance:balance, Meta:Meta{Limit:limit, Flags:map[string]bool{"beta":beta}, Aliases:[]string{alias}}}
	config.Aliases = append(config.Aliases, alias)
	config.Watchers = append(config.Watchers, User{Balance:balance + 1, Meta:Meta{Limit:limit + 1, Flags:map[string]bool{"watch":true}, Aliases:[]string{watcherAlias}}})
	return nil
}

func GetConfig(ctx chain.QueryContext) Config { return config }
func GetFeatureFlags(ctx chain.QueryContext) map[string]bool { return config.FeatureFlags }
func GetUsers(ctx chain.QueryContext) map[string]User { return config.Users }
func GetAliases(ctx chain.QueryContext) []string { return config.Aliases }
func GetWatchers(ctx chain.QueryContext) []User { return config.Watchers }

func GetUser(ctx chain.QueryContext, name string) (User, bool) {
	user, found := config.Users[name]
	return user, found
}
`

func TestGnoQueryPathStructResultRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(200),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(201),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetConfig", map[string]string{})
	expected := map[string]interface{}{
		"Owner": sender.String(),
		"FeatureFlags": map[string]interface{}{
			"alpha": true,
			"beta":  false,
		},
		"Users": map[string]interface{}{
			"alice": map[string]interface{}{
				"Balance": int64(1),
				"Meta": map[string]interface{}{
					"Limit":   int64(10),
					"Flags":   map[string]interface{}{"vip": true},
					"Aliases": []interface{}{"a1"},
				},
			},
			"bob": map[string]interface{}{
				"Balance": int64(7),
				"Meta": map[string]interface{}{
					"Limit":   int64(77),
					"Flags":   map[string]interface{}{"beta": false},
					"Aliases": []interface{}{"b1"},
				},
			},
		},
		"Aliases": []interface{}{"root", "b1"},
		"Watchers": []interface{}{
			map[string]interface{}{
				"Balance": int64(2),
				"Meta": map[string]interface{}{
					"Limit":   int64(20),
					"Flags":   map[string]interface{}{"watch": true},
					"Aliases": []interface{}{"w1"},
				},
			},
			map[string]interface{}{
				"Balance": int64(8),
				"Meta": map[string]interface{}{
					"Limit":   int64(78),
					"Flags":   map[string]interface{}{"watch": true},
					"Aliases": []interface{}{"wb1"},
				},
			},
		},
	}
	assertDeepEqualResult(t, qr.Result, expected)
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMapScalarResultRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(210),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(211),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetFeatureFlags", map[string]string{})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"alpha": true,
		"beta":  false,
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMapNamedStructResultRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(220),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(221),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetUsers", map[string]string{})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"alice": map[string]interface{}{
			"Balance": int64(1),
			"Meta": map[string]interface{}{
				"Limit":   int64(10),
				"Flags":   map[string]interface{}{"vip": true},
				"Aliases": []interface{}{"a1"},
			},
		},
		"bob": map[string]interface{}{
			"Balance": int64(7),
			"Meta": map[string]interface{}{
				"Limit":   int64(77),
				"Flags":   map[string]interface{}{"beta": false},
				"Aliases": []interface{}{"b1"},
			},
		},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathSliceScalarResultRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(230),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(231),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetAliases", map[string]string{})
	assertDeepEqualResult(t, qr.Result, []interface{}{"root", "b1"})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathSliceNamedStructResultRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(240),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(241),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetWatchers", map[string]string{})
	assertDeepEqualResult(t, qr.Result, []interface{}{
		map[string]interface{}{
			"Balance": int64(2),
			"Meta": map[string]interface{}{
				"Limit":   int64(20),
				"Flags":   map[string]interface{}{"watch": true},
				"Aliases": []interface{}{"w1"},
			},
		},
		map[string]interface{}{
			"Balance": int64(8),
			"Meta": map[string]interface{}{
				"Limit":   int64(78),
				"Flags":   map[string]interface{}{"watch": true},
				"Aliases": []interface{}{"wb1"},
			},
		},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathComplexResultWithBoolRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryResultContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(250),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryResultContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(251),
				Function: "UpsertUser",
				CallData: map[string]string{
					"name":         "bob",
					"balance":      "7",
					"limit":        "77",
					"beta":         "false",
					"alias":        "b1",
					"watcherAlias": "wb1",
				},
				ContractCode: complexQueryResultContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetUser", map[string]string{"name": "bob"})
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetUser(bob) ok=true, got %#v", qr.Ok)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Balance": int64(7),
		"Meta": map[string]interface{}{
			"Limit":   int64(77),
			"Flags":   map[string]interface{}{"beta": false},
			"Aliases": []interface{}{"b1"},
		},
	})

	qr = mustQueryContract(t, states, contract, sender, complexQueryResultContractSource, "GetUser", map[string]string{"name": "nobody"})
	if qr.Ok == nil || *qr.Ok {
		t.Fatalf("expected GetUser(nobody) ok=false, got %#v", qr.Ok)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Balance": int64(0),
		"Meta": map[string]interface{}{
			"Limit":   int64(0),
			"Flags":   nil,
			"Aliases": nil,
		},
	})

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func mustQueryContract(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	source string,
	function string,
	callData map[string]string,
) QueryResult {
	t.Helper()

	engine := NewGnoEngine()
	getStateFunc := stateGetter(states)
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: source,
		Function:     function,
		CallData:     callData,
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}

	return qr
}

func assertDeepEqualResult(t *testing.T, got interface{}, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected query result\nwant: %#v\ngot:  %#v", want, got)
	}
}
