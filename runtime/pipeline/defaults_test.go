package pipeline

import (
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/util/ps"
)

func TestDefaultImportPipelineUsesSmartContractHinterStep(t *testing.T) {
	pps := DefaultImportPS()
	if !containsProcessName(pps.Verbose(), "add-hinters") {
		t.Fatalf("import pipeline does not contain hinter hook: %v", pps.Verbose())
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
