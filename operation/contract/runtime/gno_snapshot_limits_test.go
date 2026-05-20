package runtime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const snapshotLimitMapContractSource = `package contract
import "mitum/chain"

var values map[string]string

func Initialize(ctx chain.ContractContext) error {
	values = map[string]string{}
	return nil
}

func Fill(ctx chain.ContractContext) error {
	values = map[string]string{}
	for i := 0; i < 257; i++ {
		values[string(i)] = "v"
	}
	return nil
}

func Count(ctx chain.ContractContext) int { return len(values) }
`

const snapshotLimitSliceContractSource = `package contract
import "mitum/chain"

var values []string

func Initialize(ctx chain.ContractContext) error {
	values = []string{}
	return nil
}

func Fill(ctx chain.ContractContext) error {
	values = []string{}
	for i := 0; i < 257; i++ {
		values = append(values, "v")
	}
	return nil
}

func Count(ctx chain.ContractContext) int { return len(values) }
`

const snapshotLimitBytesContractSource = `package contract
import "mitum/chain"

var blob string

func Initialize(ctx chain.ContractContext) error { return nil }

func Grow(ctx chain.ContractContext) error {
	blob = "x"
	for i := 0; i < 18; i++ {
		blob += blob
	}
	return nil
}

func Size(ctx chain.ContractContext) int { return len(blob) }
`

func TestSnapshotLimitsAllowNormalState(t *testing.T) {
	doc := SnapshotDoc{
		Version: GnoSnapshotVersion,
		Bindings: []SnapshotBinding{
			{
				Name: "values",
				Value: SnapshotValue{
					Kind: string(TypeMap),
					Entries: []SnapshotMapEntry{
						{Key: "a", Value: scalarSnapshotValue("1")},
						{Key: "b", Value: scalarSnapshotValue("2")},
					},
				},
			},
			{
				Name: "items",
				Value: SnapshotValue{
					Kind: string(TypeSlice),
					Items: []SnapshotValue{
						scalarSnapshotValue("x"),
						scalarSnapshotValue("y"),
					},
				},
			},
		},
	}

	if err := ValidateSnapshotLimits(doc, mustMarshalSnapshotDoc(t, doc)); err != nil {
		t.Fatalf("expected normal snapshot to pass limits, got: %v", err)
	}
}

func TestSnapshotLimitBoundaries(t *testing.T) {
	mapBoundary := SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: mapSnapshotValue(MaxContractSnapshotMapEntries)}},
	}
	if err := ValidateSnapshotLimits(mapBoundary, mustMarshalSnapshotDoc(t, mapBoundary)); err != nil {
		t.Fatalf("expected map entry boundary to pass, got: %v", err)
	}

	mapExceeded := SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: mapSnapshotValue(MaxContractSnapshotMapEntries + 1)}},
	}
	if err := ValidateSnapshotLimits(mapExceeded, mustMarshalSnapshotDoc(t, mapExceeded)); err == nil ||
		!strings.Contains(err.Error(), "snapshot exceeds max map entries") {
		t.Fatalf("expected map entry limit error, got: %v", err)
	}

	sliceBoundary := SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: sliceSnapshotValue(MaxContractSnapshotSliceItems)}},
	}
	if err := ValidateSnapshotLimits(sliceBoundary, mustMarshalSnapshotDoc(t, sliceBoundary)); err != nil {
		t.Fatalf("expected slice item boundary to pass, got: %v", err)
	}

	sliceExceeded := SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: sliceSnapshotValue(MaxContractSnapshotSliceItems + 1)}},
	}
	if err := ValidateSnapshotLimits(sliceExceeded, mustMarshalSnapshotDoc(t, sliceExceeded)); err == nil ||
		!strings.Contains(err.Error(), "snapshot exceeds max slice items") {
		t.Fatalf("expected slice item limit error, got: %v", err)
	}
}

func TestSnapshotBytesLimitRejectedBeforeDecode(t *testing.T) {
	snapshot := []byte(`{"version":1,"bindings":[{"name":"blob","value":{"kind":"scalar","scalar":"` +
		strings.Repeat("x", MaxContractSnapshotBytes) +
		`"}}]}`)

	if err := validateSnapshotSizeLimit(snapshot); err == nil ||
		!strings.Contains(err.Error(), "snapshot exceeds max size") {
		t.Fatalf("expected snapshot size limit error, got: %v", err)
	}
}

func TestSnapshotStatsDeterministicAcrossMapOrder(t *testing.T) {
	first := SnapshotDoc{
		Version: GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{
			Name: "values",
			Value: SnapshotValue{
				Kind: string(TypeMap),
				Entries: []SnapshotMapEntry{
					{Key: "b", Value: scalarSnapshotValue("2")},
					{Key: "a", Value: scalarSnapshotValue("1")},
				},
			},
		}},
	}
	second := SnapshotDoc{
		Version: GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{
			Name: "values",
			Value: SnapshotValue{
				Kind: string(TypeMap),
				Entries: []SnapshotMapEntry{
					{Key: "a", Value: scalarSnapshotValue("1")},
					{Key: "b", Value: scalarSnapshotValue("2")},
				},
			},
		}},
	}

	firstStats := SnapshotStatsForDoc(first, mustMarshalSnapshotDoc(t, first))
	secondStats := SnapshotStatsForDoc(second, mustMarshalSnapshotDoc(t, second))
	if firstStats != secondStats {
		t.Fatalf("expected deterministic stats for same logical map state:\nfirst:  %#v\nsecond: %#v", firstStats, secondStats)
	}
}

func TestExecuteContractFailsWhenCapturedSnapshotExceedsBytesLimit(t *testing.T) {
	engine := NewGnoEngine()
	schema := mustAnalyzeSnapshotLimitSchema(t, snapshotLimitBytesContractSource)
	contract := base.NewStringAddress("contractslimit001")
	sender := base.NewStringAddress("senderslimit001")

	states := registerSnapshotLimitContract(t, engine, schema, contract, sender, snapshotLimitBytesContractSource)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(902),
		ContractCode: snapshotLimitBytesContractSource,
		Schema:       &schema,
		Function:     "Grow",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected oversized captured snapshot error")
	}
	if !strings.Contains(err.Error(), "snapshot exceeds max size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContractFailsWhenMapEntriesExceedLimit(t *testing.T) {
	engine := NewGnoEngine()
	schema := mustAnalyzeSnapshotLimitSchema(t, snapshotLimitMapContractSource)
	contract := base.NewStringAddress("contractslimit002")
	sender := base.NewStringAddress("senderslimit002")

	states := registerSnapshotLimitContract(t, engine, schema, contract, sender, snapshotLimitMapContractSource)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(912),
		ContractCode: snapshotLimitMapContractSource,
		Schema:       &schema,
		Function:     "Fill",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected map entry limit error")
	}
	if !strings.Contains(err.Error(), "snapshot exceeds max map entries") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContractFailsWhenSliceItemsExceedLimit(t *testing.T) {
	engine := NewGnoEngine()
	schema := mustAnalyzeSnapshotLimitSchema(t, snapshotLimitSliceContractSource)
	contract := base.NewStringAddress("contractslimit003")
	sender := base.NewStringAddress("senderslimit003")

	states := registerSnapshotLimitContract(t, engine, schema, contract, sender, snapshotLimitSliceContractSource)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(922),
		ContractCode: snapshotLimitSliceContractSource,
		Schema:       &schema,
		Function:     "Fill",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected slice item limit error")
	}
	if !strings.Contains(err.Error(), "snapshot exceeds max slice items") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueryContractFailsOnExistingOversizedMapSnapshot(t *testing.T) {
	engine := NewGnoEngine()
	schema := mustAnalyzeSnapshotLimitSchema(t, snapshotLimitMapContractSource)
	contract := base.NewStringAddress("contractslimit004")
	sender := base.NewStringAddress("senderslimit004")
	states := oversizedSnapshotStates(t, contract, snapshotLimitMapContractSource, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: mapSnapshotValue(MaxContractSnapshotMapEntries + 1)}},
	})

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(932),
		ContractCode: snapshotLimitMapContractSource,
		Schema:       &schema,
		Function:     "Count",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected oversized existing map snapshot query error")
	}
	if !strings.Contains(err.Error(), "snapshot exceeds max map entries") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueryContractFailsOnExistingOversizedSliceSnapshot(t *testing.T) {
	engine := NewGnoEngine()
	schema := mustAnalyzeSnapshotLimitSchema(t, snapshotLimitSliceContractSource)
	contract := base.NewStringAddress("contractslimit005")
	sender := base.NewStringAddress("senderslimit005")
	states := oversizedSnapshotStates(t, contract, snapshotLimitSliceContractSource, SnapshotDoc{
		Version:  GnoSnapshotVersion,
		Bindings: []SnapshotBinding{{Name: "values", Value: sliceSnapshotValue(MaxContractSnapshotSliceItems + 1)}},
	})

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(942),
		ContractCode: snapshotLimitSliceContractSource,
		Schema:       &schema,
		Function:     "Count",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected oversized existing slice snapshot query error")
	}
	if !strings.Contains(err.Error(), "snapshot exceeds max slice items") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func scalarSnapshotValue(value string) SnapshotValue {
	return SnapshotValue{
		Kind:   string(TypeScalar),
		Scalar: value,
	}
}

func mapSnapshotValue(entries int) SnapshotValue {
	out := SnapshotValue{
		Kind:    string(TypeMap),
		Entries: make([]SnapshotMapEntry, entries),
	}
	for i := 0; i < entries; i++ {
		out.Entries[i] = SnapshotMapEntry{
			Key:   fmt.Sprintf("k%04d", i),
			Value: scalarSnapshotValue("v"),
		}
	}

	return out
}

func sliceSnapshotValue(items int) SnapshotValue {
	out := SnapshotValue{
		Kind:  string(TypeSlice),
		Items: make([]SnapshotValue, items),
	}
	for i := 0; i < items; i++ {
		out.Items[i] = scalarSnapshotValue("v")
	}

	return out
}

func mustAnalyzeSnapshotLimitSchema(t *testing.T, source string) ContractSchema {
	t.Helper()

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	return schema
}

func registerSnapshotLimitContract(
	t *testing.T,
	engine ContractEngine,
	schema ContractSchema,
	contract base.Address,
	sender base.Address,
	source string,
) map[string]base.State {
	t.Helper()

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(901),
		ContractCode: source,
		Schema:       &schema,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("register ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(901), result.StateMerges)
	return states
}

func oversizedSnapshotStates(
	t *testing.T,
	contract base.Address,
	source string,
	doc SnapshotDoc,
) map[string]base.State {
	t.Helper()

	return map[string]base.State{
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			base.Height(930),
			pstate.RuntimeStateKey(contract),
			deriveRuntimeState(contract, source),
			nil,
			nil,
		),
		pstate.SnapshotStateKey(contract): common.NewBaseState(
			base.Height(930),
			pstate.SnapshotStateKey(contract),
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, mustMarshalSnapshotDoc(t, doc)),
			nil,
			nil,
		),
	}
}
