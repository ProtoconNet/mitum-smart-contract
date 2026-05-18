package runtime

import (
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const complexQueryArgContractSource = `package contract
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

type UserSelector struct {
	Name string
	RequireActive bool
}

var flags map[string]bool
var users map[string]User
var aliases []string
var watchers []User

func Initialize(ctx chain.ContractContext) error {
	flags = map[string]bool{"alpha": true, "beta": false}
	users = map[string]User{
		"alice": User{Balance: 10, Meta: Meta{Limit: 100, Flags: map[string]bool{"active": true}, Aliases: []string{"a1"}}},
		"bob": User{Balance: 20, Meta: Meta{Limit: 200, Flags: map[string]bool{"active": false}, Aliases: []string{"b1"}}},
	}
	aliases = []string{"root", "child"}
	watchers = []User{
		User{Balance: 30, Meta: Meta{Limit: 300, Flags: map[string]bool{"active": true}, Aliases: []string{"w1"}}},
		User{Balance: 40, Meta: Meta{Limit: 400, Flags: map[string]bool{"active": false}, Aliases: []string{"w2"}}},
	}
	return nil
}

func GetSelectedUser(ctx chain.ContractContext, selector UserSelector) (User, bool) {
	user, found := users[selector.Name]
	if !found {
		return User{}, false
	}
	if selector.RequireActive && !user.Meta.Flags["active"] {
		return User{}, false
	}
	return user, true
}

func EchoFlags(ctx chain.ContractContext, next map[string]bool) map[string]bool {
	return next
}

func EchoUsers(ctx chain.ContractContext, next map[string]User) map[string]User {
	return next
}

func EchoAliases(ctx chain.ContractContext, next []string) []string {
	return next
}

func EchoWatchers(ctx chain.ContractContext, next []User) []User {
	return next
}

func HasAlias(ctx chain.ContractContext, name string) bool {
	for _, alias := range aliases {
		if alias == name {
			return true
		}
	}
	return false
}
`

func TestGnoQueryPathStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(400),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "GetSelectedUser", map[string]string{
		"selector": `{"name":"alice","requireactive":true}`,
	})
	if qr.Ok == nil || !*qr.Ok {
		t.Fatalf("expected GetSelectedUser ok=true, got %#v", qr.Ok)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Balance": int64(10),
		"Meta": map[string]interface{}{
			"Limit":   int64(100),
			"Flags":   map[string]interface{}{"active": true},
			"Aliases": []interface{}{"a1"},
		},
	})

	qr = mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "GetSelectedUser", map[string]string{
		"selector": `{"name":"bob","requireactive":true}`,
	})
	if qr.Ok == nil || *qr.Ok {
		t.Fatalf("expected GetSelectedUser ok=false, got %#v", qr.Ok)
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

func TestGnoQueryPathMapScalarArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(410),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoFlags", map[string]string{
		"next": `{"beta":false,"alpha":true}`,
	})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{"alpha": true, "beta": false})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMapNamedStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(420),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoUsers", map[string]string{
		"next": `{"bob":{"balance":7,"meta":{"limit":77,"flags":{"vip":true},"aliases":["b1"]}}}`,
	})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"bob": map[string]interface{}{
			"Balance": int64(7),
			"Meta": map[string]interface{}{
				"Limit":   int64(77),
				"Flags":   map[string]interface{}{"vip": true},
				"Aliases": []interface{}{"b1"},
			},
		},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathSliceScalarArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(430),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoAliases", map[string]string{
		"next": `["a1","a2","a3"]`,
	})
	assertDeepEqualResult(t, qr.Result, []interface{}{"a1", "a2", "a3"})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathSliceNamedStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(440),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoWatchers", map[string]string{
		"next": `[{"balance":9,"meta":{"limit":99,"flags":{"watch":true},"aliases":["w1"]}},{"balance":10,"meta":{"limit":100,"flags":{"watch":false},"aliases":["w2"]}}]`,
	})
	assertDeepEqualResult(t, qr.Result, []interface{}{
		map[string]interface{}{
			"Balance": int64(9),
			"Meta": map[string]interface{}{
				"Limit":   int64(99),
				"Flags":   map[string]interface{}{"watch": true},
				"Aliases": []interface{}{"w1"},
			},
		},
		map[string]interface{}{
			"Balance": int64(10),
			"Meta": map[string]interface{}{
				"Limit":   int64(100),
				"Flags":   map[string]interface{}{"watch": false},
				"Aliases": []interface{}{"w2"},
			},
		},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathNilEmptyCompositeArgsPolicy(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(450),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoFlags", map[string]string{
		"next": `null`,
	})
	if qr.Result != nil {
		t.Fatalf("expected nil map query arg result, got %#v", qr.Result)
	}

	qr = mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoFlags", map[string]string{
		"next": `{}`,
	})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{})

	qr = mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoAliases", map[string]string{
		"next": `null`,
	})
	if qr.Result != nil {
		t.Fatalf("expected nil slice query arg result, got %#v", qr.Result)
	}

	qr = mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "EchoAliases", map[string]string{
		"next": `[]`,
	})
	assertDeepEqualResult(t, qr.Result, []interface{}{})

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoQueryPathMalformedCompositeArgRejected(t *testing.T) {
	_, states, contract, sender, _ := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(460),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: complexQueryArgContractSource,
		Function:     "EchoAliases",
		CallData:     map[string]string{"next": `["a1",}`},
	})
	if err == nil {
		t.Fatalf("expected malformed composite query arg to fail")
	}
}

func TestGnoQueryPathScalarArgPlainStringRegression(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexQueryArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(470),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexQueryArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexQueryArgContractSource, "HasAlias", map[string]string{
		"name": "root",
	})
	if got := qr.Result.(bool); !got {
		t.Fatalf("expected HasAlias(root)=true")
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
