package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const complexWriteArgContractSource = `package contract
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
var flags map[string]bool
var users map[string]User
var aliases []string
var watchers []User
var label string

func Initialize(ctx chain.ContractContext) error { return nil }

func SetConfig(ctx chain.ContractContext, cfg Config) error {
	config = cfg
	return nil
}

func SetFlags(ctx chain.ContractContext, next map[string]bool) error {
	flags = next
	return nil
}

func ReplaceUsers(ctx chain.ContractContext, next map[string]User) error {
	users = next
	return nil
}

func SetAliases(ctx chain.ContractContext, next []string) error {
	aliases = next
	return nil
}

func ReplaceWatchers(ctx chain.ContractContext, next []User) error {
	watchers = next
	return nil
}

func SetLabel(ctx chain.ContractContext, next string) error {
	label = next
	return nil
}

func GetConfig(ctx chain.ContractContext) Config { return config }
func GetFlags(ctx chain.ContractContext) map[string]bool { return flags }
func GetUsers(ctx chain.ContractContext) map[string]User { return users }
func GetAliases(ctx chain.ContractContext) []string { return aliases }
func GetWatchers(ctx chain.ContractContext) []User { return watchers }
func GetLabel(ctx chain.ContractContext) string { return label }
`

func TestGnoWritePathStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(300),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:     InvocationModeCall,
				Height:   base.Height(301),
				Function: "SetConfig",
				CallData: map[string]string{
					"cfg": `{"owner":"alice","featureflags":{"alpha":true},"users":{"bob":{"balance":7,"meta":{"limit":77,"flags":{"vip":true},"aliases":["b1"]}}},"aliases":["a1","a2"],"watchers":[{"balance":8,"meta":{"limit":88,"flags":{"watch":true},"aliases":["w1"]}}]}`,
				},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetConfig", map[string]string{})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Owner":        "alice",
		"FeatureFlags": map[string]interface{}{"alpha": true},
		"Users": map[string]interface{}{
			"bob": map[string]interface{}{
				"Balance": int64(7),
				"Meta": map[string]interface{}{
					"Limit":   int64(77),
					"Flags":   map[string]interface{}{"vip": true},
					"Aliases": []interface{}{"b1"},
				},
			},
		},
		"Aliases": []interface{}{"a1", "a2"},
		"Watchers": []interface{}{
			map[string]interface{}{
				"Balance": int64(8),
				"Meta": map[string]interface{}{
					"Limit":   int64(88),
					"Flags":   map[string]interface{}{"watch": true},
					"Aliases": []interface{}{"w1"},
				},
			},
		},
	})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoWritePathMapScalarArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(310),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(311),
				Function:     "SetFlags",
				CallData:     map[string]string{"next": `{"beta":false,"alpha":true}`},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetFlags", map[string]string{})
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{"alpha": true, "beta": false})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoWritePathMapNamedStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(320),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(321),
				Function:     "ReplaceUsers",
				CallData:     map[string]string{"next": `{"bob":{"balance":7,"meta":{"limit":77,"flags":{"vip":true},"aliases":["b1"]}}}`},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetUsers", map[string]string{})
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

func TestGnoWritePathSliceScalarArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(330),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(331),
				Function:     "SetAliases",
				CallData:     map[string]string{"next": `["a1","a2","a3"]`},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetAliases", map[string]string{})
	assertDeepEqualResult(t, qr.Result, []interface{}{"a1", "a2", "a3"})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoWritePathSliceNamedStructArgRoundTrip(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(340),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(341),
				Function:     "ReplaceWatchers",
				CallData:     map[string]string{"next": `[{"balance":9,"meta":{"limit":99,"flags":{"watch":true},"aliases":["w1"]}},{"balance":10,"meta":{"limit":100,"flags":{"watch":false},"aliases":["w2"]}}]`},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetWatchers", map[string]string{})
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

func TestGnoWritePathNilEmptyArgPolicy(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(350),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(351),
				Function:     "SetFlags",
				CallData:     map[string]string{"next": `null`},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(352),
				Function:     "SetAliases",
				CallData:     map[string]string{"next": `[]`},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetFlags", map[string]string{})
	if qr.Result != nil {
		t.Fatalf("expected nil map result, got %#v", qr.Result)
	}

	qr = mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetAliases", map[string]string{})
	assertDeepEqualResult(t, qr.Result, []interface{}{})
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func TestGnoWritePathMalformedJSONArgRejected(t *testing.T) {
	engine := NewGnoEngine()
	encs := newRuntimeTestEncoders(t)

	contract := base.NewStringAddress("contract0011")
	sender := base.NewStringAddress("sender0011")
	getStateFunc := func(key string) (base.State, bool, error) { return nil, false, nil }

	_, err := engine.ExecuteContract(encs, getStateFunc, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(360),
		ContractCode: complexWriteArgContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	runtimeState := deriveRuntimeState(contract, complexWriteArgContractSource)
	snapshotState := pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, nil)
	states := map[string]base.State{
		pstate.RuntimeStateKey(contract):  runtimeStateState(base.Height(360), contract, runtimeState),
		pstate.SnapshotStateKey(contract): snapshotStateState(base.Height(360), contract, snapshotState),
	}

	_, err = engine.ExecuteContract(encs, stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(361),
		ContractCode: complexWriteArgContractSource,
		Function:     "SetAliases",
		CallData:     map[string]string{"next": `["ok",}`},
	})
	if err == nil {
		t.Fatalf("expected malformed JSON arg error")
	}
}

func TestGnoWritePathScalarArgPlainStringStillWorks(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		complexWriteArgContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(370),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: complexWriteArgContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(371),
				Function:     "SetLabel",
				CallData:     map[string]string{"next": "hello"},
				ContractCode: complexWriteArgContractSource,
			},
		},
	)

	qr := mustQueryContract(t, states, contract, sender, complexWriteArgContractSource, "GetLabel", map[string]string{})
	if got := qr.Result.(string); got != "hello" {
		t.Fatalf("unexpected GetLabel result: %s", got)
	}
	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}

func runtimeStateState(height base.Height, contract base.Address, value pstate.RuntimeStateValue) base.State {
	return common.NewBaseState(height, pstate.RuntimeStateKey(contract), value, nil, nil)
}

func snapshotStateState(height base.Height, contract base.Address, value pstate.SnapshotStateValue) base.State {
	return common.NewBaseState(height, pstate.SnapshotStateKey(contract), value, nil, nil)
}
