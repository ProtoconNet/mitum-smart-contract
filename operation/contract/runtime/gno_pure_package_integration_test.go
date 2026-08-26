package runtime

import (
	"testing"

	"github.com/imfact-labs/mitum2/base"
)

const uint256PurePackageSmokeContract = `package contract
import (
	"mitum/math/v1/u256"
	"mitum/chain"
)

var value uint64

func Initialize(ctx chain.WriteContext) error {
	value = u256.One().Uint64()
	return nil
}

func Store(ctx chain.WriteContext, input uint64) error {
	value = u256.NewUint(input).Uint64()
	return nil
}

func Get(ctx chain.QueryContext) uint64 {
	return u256.NewUint(value).Uint64()
}
`

func TestOnblocUint256PurePackageWriteAndQuerySmoke(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractuint256smoke")
	sender := base.NewStringAddress("senderuint256smoke01")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode: InvocationModeRegister, Contract: contract, Sender: sender,
		Height: base.Height(1), ContractCode: uint256PurePackageSmokeContract,
	})
	if err != nil {
		t.Fatalf("register uint256 contract: %v", err)
	}
	applyStateMerges(states, base.Height(1), result.StateMerges)

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode: InvocationModeCall, Contract: contract, Sender: sender,
		Height: base.Height(2), ContractCode: uint256PurePackageSmokeContract,
		Function: "Store", CallData: map[string]string{"input": "7"},
	})
	if err != nil {
		t.Fatalf("write using uint256 package: %v", err)
	}
	applyStateMerges(states, base.Height(2), result.StateMerges)

	query, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract: contract, Sender: sender, Height: base.Height(2),
		ContractCode: uint256PurePackageSmokeContract, Function: "Get", CallData: map[string]string{},
	})
	if err != nil {
		t.Fatalf("query using uint256 package: %v", err)
	}
	if got, ok := query.Result.(uint64); !ok || got != 7 {
		t.Fatalf("expected uint256 smoke result 7, got %#v", query.Result)
	}
}
