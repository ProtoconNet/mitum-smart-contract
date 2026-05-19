package contract

import (
	"fmt"
	"strings"
	"testing"

	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	types "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

func TestRegisterContractFactInitDataPayloadLimits(t *testing.T) {
	tests := []struct {
		name      string
		callData  map[string]string
		wantError string
	}{
		{
			name:     "within limit",
			callData: map[string]string{"initialValue": "seed", "initialLimit": "10"},
		},
		{
			name:      "entry count exceeded",
			callData:  contractPayloadEntries(cruntime.MaxContractCallDataEntries + 1),
			wantError: "max entries",
		},
		{
			name:      "key size exceeded",
			callData:  map[string]string{strings.Repeat("k", cruntime.MaxContractCallDataKeyBytes+1): "v"},
			wantError: "key exceeds max size",
		},
		{
			name:      "value size exceeded",
			callData:  map[string]string{"value": strings.Repeat("v", cruntime.MaxContractCallDataValueBytes+1)},
			wantError: "value for key",
		},
		{
			name:      "total size exceeded",
			callData:  contractPayloadTotalBytesExceeded(),
			wantError: "max total key+value size",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fact := newRegisterPayloadLimitFact(i, tt.callData)
			err := fact.IsValid(nil)
			assertPayloadLimitValidation(t, err, tt.wantError)
		})
	}
}

func TestCallContractFactCallDataPayloadLimits(t *testing.T) {
	tests := []struct {
		name      string
		callData  map[string]string
		wantError string
	}{
		{
			name:     "within limit",
			callData: map[string]string{"function": "UpdateData", "value": "next"},
		},
		{
			name:      "entry count exceeded",
			callData:  contractPayloadEntries(cruntime.MaxContractCallDataEntries + 1),
			wantError: "max entries",
		},
		{
			name:      "key size exceeded",
			callData:  map[string]string{strings.Repeat("k", cruntime.MaxContractCallDataKeyBytes+1): "v"},
			wantError: "key exceeds max size",
		},
		{
			name:      "value size exceeded",
			callData:  map[string]string{"value": strings.Repeat("v", cruntime.MaxContractCallDataValueBytes+1)},
			wantError: "value for key",
		},
		{
			name:      "total size exceeded",
			callData:  contractPayloadTotalBytesExceeded(),
			wantError: "max total key+value size",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fact := newCallPayloadLimitFact(i, tt.callData)
			err := fact.IsValid(nil)
			assertPayloadLimitValidation(t, err, tt.wantError)
		})
	}
}

func newRegisterPayloadLimitFact(i int, callData map[string]string) RegisterContractFact {
	return NewRegisterContractFact(
		[]byte("token"),
		base.NewStringAddress(fmt.Sprintf("senderpayload%04d", i)),
		base.NewStringAddress(fmt.Sprintf("contractpayload%04d", i)),
		"package contract\n",
		callData,
		types.CurrencyID("ABC"),
	)
}

func newCallPayloadLimitFact(i int, callData map[string]string) CallContractFact {
	return NewCallContractFact(
		[]byte("token"),
		base.NewStringAddress(fmt.Sprintf("sendercall%04d", i)),
		base.NewStringAddress(fmt.Sprintf("contractcall%04d", i)),
		callData,
		types.CurrencyID("ABC"),
	)
}

func contractPayloadEntries(count int) map[string]string {
	out := make(map[string]string, count)
	for i := 0; i < count; i++ {
		out[fmt.Sprintf("k%02d", i)] = "v"
	}

	return out
}

func contractPayloadTotalBytesExceeded() map[string]string {
	out := make(map[string]string, cruntime.MaxContractCallDataEntries)
	for i := 0; i < cruntime.MaxContractCallDataEntries; i++ {
		out[fmt.Sprintf("k%02d", i)] = strings.Repeat("v", 1030)
	}

	return out
}

func assertPayloadLimitValidation(t *testing.T, err error, wantError string) {
	t.Helper()

	if wantError == "" {
		if err != nil {
			t.Fatalf("IsValid returned error: %v", err)
		}
		return
	}

	if err == nil {
		t.Fatalf("expected error containing %q", wantError)
	}
	if !strings.Contains(err.Error(), wantError) {
		t.Fatalf("expected error containing %q, got: %v", wantError, err)
	}
}
