package runtime

import (
	"encoding/json"
	"testing"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

const flatStructEngineContractSource = `package contract
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
`

func TestGnoWritePathFlatStructRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(flatStructEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0001")
	sender := base.NewStringAddress("sender0001")
	runtimeValue := deriveRuntimeState(contract, flatStructEngineContractSource)
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
			Height:       base.Height(10),
			ContractCode: flatStructEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertFlatStructBinding(t, registerSnapshot, "config", map[string]string{
		"Owner":  sender.String(),
		"Paused": "false",
		"Limit":  "1",
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
			Height:       base.Height(11),
			ContractCode: flatStructEngineContractSource,
			Function:     "UpdateConfig",
			CallData: map[string]string{
				"paused": "true",
				"limit":  "7",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertFlatStructBinding(t, callSnapshot, "config", map[string]string{
		"Owner":  sender.String(),
		"Paused": "true",
		"Limit":  "7",
	})

	if string(mustMarshalSnapshotDoc(t, registerSnapshot)) == string(mustMarshalSnapshotDoc(t, callSnapshot)) {
		t.Fatalf("expected call snapshot to differ after state mutation")
	}
}

func executeGnoWriteForTest(
	t *testing.T,
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	schema ContractSchema,
	packagePath string,
	req ExecuteRequest,
	snapshot []byte,
) SnapshotDoc {
	t.Helper()

	execCtx, err := NewExecutionContext(
		encs,
		getStateFunc,
		req.Contract,
		req.Sender,
		req.Height,
		false,
	)
	if err != nil {
		t.Fatalf("NewExecutionContext returned error: %v", err)
	}

	limits := WriteGnoExecutionLimits()
	gasMeter := NewGnoGasMeter(limits.GasLimit)
	m, pkg, err := newGnoMachineAndPackage(
		execCtx,
		packagePath,
		req.ContractCode,
		limits,
		gasMeter,
	)
	if err != nil {
		t.Fatalf("newGnoMachineAndPackage returned error: %v", err)
	}

	if err := RestoreSnapshot(m, pkg, snapshot, schema); err != nil {
		t.Fatalf("RestoreSnapshot returned error: %v", err)
	}

	if err := invokeTypedWrite(m, pkg, req, schema); err != nil {
		t.Fatalf("invokeTypedWrite returned error: %v", err)
	}

	snapshotBytes, err := CaptureSnapshot(pkg, m.Store, schema)
	if err != nil {
		t.Fatalf("CaptureSnapshot returned error: %v", err)
	}

	var doc SnapshotDoc
	if err := json.Unmarshal(snapshotBytes, &doc); err != nil {
		t.Fatalf("json.Unmarshal(snapshot) returned error: %v", err)
	}

	return doc
}

func newRuntimeTestEncoders(t *testing.T) encoder.Encoders {
	t.Helper()

	enc := jsonenc.NewEncoder()
	encs := encoder.NewEncoders(enc, enc)
	if err := encs.AddDetail(encoder.DecodeDetail{
		Hint:     base.StringAddressHint,
		Instance: base.StringAddress{},
	}); err != nil {
		t.Fatalf("AddDetail(StringAddress) returned error: %v", err)
	}

	return *encs
}

func assertFlatStructBinding(t *testing.T, doc SnapshotDoc, name string, expected map[string]string) {
	t.Helper()

	for _, binding := range doc.Bindings {
		if binding.Name != name {
			continue
		}

		if binding.Value.Kind != string(TypeStruct) {
			t.Fatalf("expected struct snapshot binding, got %#v", binding.Value)
		}

		if len(binding.Value.Fields) != len(expected) {
			t.Fatalf("unexpected field count: %#v", binding.Value.Fields)
		}

		for _, field := range binding.Value.Fields {
			want, found := expected[field.Name]
			if !found {
				t.Fatalf("unexpected field %q in snapshot", field.Name)
			}
			if field.Value.Scalar != want {
				t.Fatalf("unexpected value for field %q: want %q, got %#v", field.Name, want, field.Value)
			}
		}

		return
	}

	t.Fatalf("binding %q not found in snapshot doc", name)
}
