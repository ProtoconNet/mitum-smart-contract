package cmds

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	contractop "github.com/ProtoconNet/mitum-smart-contract/operation/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/alecthomas/kong"
)

func TestCallContractCommandKongSchemaAcceptsBatchFlags(t *testing.T) {
	var cli struct {
		Call CallContractCommand `cmd:"" name:"call"`
	}

	if _, err := kong.New(&cli, kong.Vars{"network_id": "mitum"}); err != nil {
		t.Fatalf("kong.New returned error: %v", err)
	}
}

func TestNewCallContractFactFromCLIInputLegacyCallData(t *testing.T) {
	fact, err := testCallContractFactFromCLI(
		t,
		ptrString(`{"function":"UpdateData","value":"next"}`),
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("newCallContractFactFromCLIInput returned error: %v", err)
	}

	assertCLIItems(t, fact.Items(), []contractop.CallContractItem{
		contractop.NewCallContractItem("UpdateData", map[string]string{"value": "next"}),
	})
	if got := fact.CallData(); got["function"] != "UpdateData" || got["value"] != "next" {
		t.Fatalf("legacy CallData view mismatch: %#v", got)
	}
}

func TestNewCallContractFactFromCLIInputItems(t *testing.T) {
	fact, err := testCallContractFactFromCLI(
		t,
		nil,
		ptrString(`[
			{"function":"CreateData","call_data":{"id":"a","value":"one"}},
			{"function":"UpdateData","call_data":{"id":"a","value":"two"}}
		]`),
		nil,
	)
	if err != nil {
		t.Fatalf("newCallContractFactFromCLIInput returned error: %v", err)
	}

	assertCLIItems(t, fact.Items(), testCLIBatchItems())
	if got := fact.CallData(); got != nil {
		t.Fatalf("multi-item fact must not expose legacy CallData view: %#v", got)
	}
}

func TestNewCallContractFactFromCLIInputItemsFile(t *testing.T) {
	itemsFile := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(itemsFile, []byte(`[
		{"function":"CreateData","call_data":{"id":"a","value":"one"}},
		{"function":"UpdateData","call_data":{"id":"a","value":"two"}}
	]`), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	fact, err := testCallContractFactFromCLI(t, nil, nil, &itemsFile)
	if err != nil {
		t.Fatalf("newCallContractFactFromCLIInput returned error: %v", err)
	}

	assertCLIItems(t, fact.Items(), testCLIBatchItems())
}

func TestNewCallContractFactFromCLIInputSingleItemItemsAllowed(t *testing.T) {
	fact, err := testCallContractFactFromCLI(
		t,
		nil,
		ptrString(`[{"function":"UpdateData","call_data":{"value":"next"}}]`),
		nil,
	)
	if err != nil {
		t.Fatalf("newCallContractFactFromCLIInput returned error: %v", err)
	}

	assertCLIItems(t, fact.Items(), []contractop.CallContractItem{
		contractop.NewCallContractItem("UpdateData", map[string]string{"value": "next"}),
	})
	if got := fact.CallData(); got["function"] != "UpdateData" || got["value"] != "next" {
		t.Fatalf("single item --items should have legacy CallData view: %#v", got)
	}
}

func TestNewCallContractFactFromCLIInputRequiresExactlyOneInput(t *testing.T) {
	tests := []struct {
		name      string
		callData  *string
		items     *string
		itemsFile *string
	}{
		{name: "none"},
		{
			name:     "calldata and items",
			callData: ptrString(`{"function":"UpdateData"}`),
			items:    ptrString(`[{"function":"UpdateData","call_data":{}}]`),
		},
		{
			name:      "calldata and items file",
			callData:  ptrString(`{"function":"UpdateData"}`),
			itemsFile: ptrString("items.json"),
		},
		{
			name:      "items and items file",
			items:     ptrString(`[{"function":"UpdateData","call_data":{}}]`),
			itemsFile: ptrString("items.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testCallContractFactFromCLI(t, tt.callData, tt.items, tt.itemsFile)
			if err == nil {
				t.Fatal("expected exactly-one input error")
			}
			if !strings.Contains(err.Error(), "exactly one of --calldata, --items, or --items-file is required") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewCallContractFactFromCLIInputRejectsMalformedItems(t *testing.T) {
	_, err := testCallContractFactFromCLI(t, nil, ptrString(`[{"function":`), nil)
	if err == nil {
		t.Fatal("expected malformed --items error")
	}
	if !strings.Contains(err.Error(), "invalid --items JSON array") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCallContractFactFromCLIInputRejectsUnreadableItemsFile(t *testing.T) {
	itemsFile := filepath.Join(t.TempDir(), "missing.json")

	_, err := testCallContractFactFromCLI(t, nil, nil, &itemsFile)
	if err == nil {
		t.Fatal("expected unreadable --items-file error")
	}
	if !strings.Contains(err.Error(), "failed to read --items-file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCallContractFactFromCLIInputRejectsEmptyItemsFilePath(t *testing.T) {
	_, err := testCallContractFactFromCLI(t, nil, nil, ptrString(""))
	if err == nil {
		t.Fatal("expected empty --items-file path error")
	}
	if !strings.Contains(err.Error(), "--items-file path is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testCallContractFactFromCLI(
	t *testing.T,
	callData, items, itemsFile *string,
) (contractop.CallContractFact, error) {
	t.Helper()

	return newCallContractFactFromCLIInput(
		[]byte("cli-token"),
		base.NewStringAddress("clicallsender001"),
		base.NewStringAddress("clicallcontract1"),
		callData,
		items,
		itemsFile,
		ctypes.CurrencyID("ABC"),
	)
}

func testCLIBatchItems() []contractop.CallContractItem {
	return []contractop.CallContractItem{
		contractop.NewCallContractItem("CreateData", map[string]string{"id": "a", "value": "one"}),
		contractop.NewCallContractItem("UpdateData", map[string]string{"id": "a", "value": "two"}),
	}
}

func assertCLIItems(t *testing.T, got, want []contractop.CallContractItem) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("unexpected item count: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Function() != want[i].Function() {
			t.Fatalf("item %d function mismatch: got %q, want %q", i, got[i].Function(), want[i].Function())
		}
		gotData := got[i].CallData()
		wantData := want[i].CallData()
		if len(gotData) != len(wantData) {
			t.Fatalf("item %d call_data length mismatch: got %#v, want %#v", i, gotData, wantData)
		}
		for key, wantValue := range wantData {
			if gotData[key] != wantValue {
				t.Fatalf("item %d call_data[%q] mismatch: got %q, want %q", i, key, gotData[key], wantValue)
			}
		}
	}
}

func ptrString(s string) *string {
	return &s
}
