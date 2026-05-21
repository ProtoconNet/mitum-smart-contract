package runtime

import (
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

func TestValidateContractMalformedSourceFailsAtSchemaLayer(t *testing.T) {
	original := newGnoMachineAndPackageFunc
	machineCalls := 0
	newGnoMachineAndPackageFunc = func(
		*ExecutionContext,
		string,
		string,
		GnoExecutionLimits,
		gstore.GasMeter,
	) (*gno.Machine, *gno.PackageValue, error) {
		machineCalls++
		return nil, nil, nil
	}
	defer func() { newGnoMachineAndPackageFunc = original }()

	engine := NewGnoEngine()
	_, err := engine.ValidateContract("package contract\nfunc Initialize(")
	if err == nil {
		t.Fatal("expected malformed source error")
	}
	if !containsAll(err.Error(), "failed to analyze contract schema", "failed to parse contract source for schema analysis") {
		t.Fatalf("unexpected error: %v", err)
	}
	if machineCalls != 0 {
		t.Fatalf("expected VM path not to run, got %d calls", machineCalls)
	}
}

func TestValidateContractDisallowedImportFailsAtSchemaLayer(t *testing.T) {
	engine := NewGnoEngine()
	_, err := engine.ValidateContract(`package contract
import (
	"fmt"
	"mitum/chain"
)

func Initialize(ctx chain.WriteContext) error {
	return nil
}
`)
	if err == nil {
		t.Fatal("expected disallowed import error")
	}
	if !strings.Contains(err.Error(), `import "fmt" is not allowed`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateContractComplexityLimitFailsAtSchemaLayer(t *testing.T) {
	engine := NewGnoEngine()
	_, err := engine.ValidateContract(schemaComplexityFunctionCountSource(MaxContractSchemaFunctions + 1))
	if err == nil {
		t.Fatal("expected complexity limit error")
	}
	if !strings.Contains(err.Error(), "max function count") {
		t.Fatalf("unexpected error: %v", err)
	}
}
