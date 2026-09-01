package pipeline

import (
	"context"
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
)

func TestDefaultImportPipelineUsesSmartContractHinterStep(t *testing.T) {
	pps := DefaultImportPS()
	if !containsProcessName(pps.Verbose(), "add-hinters") {
		t.Fatalf("import pipeline does not contain hinter hook: %v", pps.Verbose())
	}
}

func TestDefaultRunPipelineIncludesSmartContractProcessorStep(t *testing.T) {
	noop := func(ctx context.Context) (context.Context, error) { return ctx, nil }
	pps := DefaultRunPS(RunHooks{
		DigesterName:             ps.Name("test-smart-digester"),
		Digester:                 noop,
		StartDigesterName:        ps.Name("test-smart-start-digester"),
		StartDigester:            noop,
		CheckHold:                noop,
		ProposalProcessors:       noop,
		WhenNewBlockSaved:        noop,
		WhenNewBlockConfirmed:    noop,
		WhenNewBlockSavedSyncing: noop,
		DigestAPIHandlers:        noop,
		DigesterFollowUp:         noop,
	})
	if !containsProcessName(pps.Verbose(), steps.PNameOperationProcessorsMap.String()) {
		t.Fatalf("run pipeline does not contain processor step: %v", pps.Verbose())
	}
}

func containsProcessName(names []ps.Name, name string) bool {
	for i := range names {
		if strings.HasPrefix(names[i].String(), name+":") || names[i].String() == name {
			return true
		}
	}
	return false
}
