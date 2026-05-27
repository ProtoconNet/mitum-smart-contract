package runtime

import (
	"encoding/hex"
	"strings"
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"golang.org/x/crypto/sha3"
)

const sha3NativeContractSource = `package contract
import "mitum/chain"

var initializedDigest string
var storedDigest string

func Initialize(ctx chain.WriteContext, value string) error {
	initializedDigest = chain.SHA3Sum256(value)
	return nil
}

func StoreDigest(ctx chain.WriteContext, data string) error {
	storedDigest = chain.SHA3Sum256(data)
	return nil
}

func GetInitializedDigest(ctx chain.QueryContext) string { return initializedDigest }
func GetStoredDigest(ctx chain.QueryContext) string { return storedDigest }
func Digest(ctx chain.QueryContext, data string) string { return chain.SHA3Sum256(data) }
`

func TestSHA3Sum256NativeCorrectnessAndRuntimeAvailability(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractsha30001")
	sender := base.NewStringAddress("sendersha300001")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: sha3NativeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{"value": "abc"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(600), result.StateMerges)

	assertSHA3Query(t, engine, states, contract, sender, "GetInitializedDigest", nil, knownSHA3ABC)
	assertSHA3Query(t, engine, states, contract, sender, "Digest", map[string]string{"data": ""}, knownSHA3Empty)
	assertSHA3Query(t, engine, states, contract, sender, "Digest", map[string]string{"data": "가"}, expectedSHA3Hex("가"))
	assertSHA3Query(t, engine, states, contract, sender, "Digest", map[string]string{"data": "ff"}, expectedSHA3Hex("ff"))

	hexDecodedFF := sha3.Sum256([]byte{0xff})
	if expectedSHA3Hex("ff") == hex.EncodeToString(hexDecodedFF[:]) {
		t.Fatal(`expected SHA3Sum256("ff") to hash raw "f","f" bytes, not hex-decoded 0xff`)
	}

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(601),
		ContractCode: sha3NativeContractSource,
		Function:     "StoreDigest",
		CallData:     map[string]string{"data": "ff"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(StoreDigest) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(601), result.StateMerges)
	assertSHA3Query(t, engine, states, contract, sender, "GetStoredDigest", nil, expectedSHA3Hex("ff"))
}

func TestSHA3Sum256NativeHasSingleStringResult(t *testing.T) {
	schema, err := AnalyzeContractSchema(sha3NativeContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	fn, found := schema.FindFunction("Digest")
	if !found {
		t.Fatal("Digest function not found")
	}
	if len(fn.Results) != 1 || fn.Results[0].Type.Kind != TypeScalar || fn.Results[0].Type.Scalar != "string" {
		t.Fatalf("expected single string result, got %#v", fn.Results)
	}
}

func TestSHA3Sum256HelperHashesInvalidUTF8RawBytes(t *testing.T) {
	raw := []byte{0xff, 0xfe, 'f'}
	data := string(raw)

	sum := sha3.Sum256(raw)
	want := hex.EncodeToString(sum[:])
	if got := sha3Sum256HexString(data); got != want {
		t.Fatalf("unexpected invalid UTF-8 digest: got %q, want %q", got, want)
	}
}

func assertSHA3Query(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	callData map[string]string,
	want string,
) {
	t.Helper()

	if callData == nil {
		callData = map[string]string{}
	}
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: sha3NativeContractSource,
		Function:     function,
		CallData:     callData,
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}
	got, ok := qr.Result.(string)
	if !ok {
		t.Fatalf("expected string result for %s, got %#v", function, qr.Result)
	}
	if got != want {
		t.Fatalf("unexpected %s digest:\ngot  %s\nwant %s", function, got, want)
	}
	if strings.ToLower(got) != got {
		t.Fatalf("expected lowercase hex digest, got %q", got)
	}
}

func expectedSHA3Hex(data string) string {
	sum := sha3.Sum256([]byte(data))

	return hex.EncodeToString(sum[:])
}

const (
	knownSHA3Empty = "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	knownSHA3ABC   = "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532"
)
