package contract

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCallContractFactDoesNotStoreLegacyCallData(t *testing.T) {
	if _, found := reflect.TypeOf(CallContractFact{}).FieldByName("legacyCallData"); found {
		t.Fatal("CallContractFact must not store legacyCallData; legacy call_data is only a compatibility view")
	}
}

func TestCallContractItemSetsHint(t *testing.T) {
	item := NewCallContractItem("UpdateData", map[string]string{"value": "next"})
	if !item.Hint().Equal(CallContractItemHint) {
		t.Fatalf("unexpected item hint: got %q, want %q", item.Hint(), CallContractItemHint)
	}
	if err := item.IsValid(nil); err != nil {
		t.Fatalf("IsValid returned error: %v", err)
	}
}

func TestCallContractItemRejectsInvalidHint(t *testing.T) {
	item := NewCallContractItem("UpdateData", map[string]string{"value": "next"})
	item.BaseHinter = hint.NewBaseHinter(hint.MustNewHint("mitum-contract-other-call-item-v0.0.1"))

	err := item.IsValid(nil)
	if err == nil {
		t.Fatal("expected invalid item hint error")
	}
	if !strings.Contains(err.Error(), "type does not match") {
		t.Fatalf("expected hint type mismatch error, got %v", err)
	}
}

func TestCallContractItemJSONHintRoundTrip(t *testing.T) {
	item := NewCallContractItem("UpdateData", map[string]string{"value": "next"})
	b, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	var gotHint string
	if err := json.Unmarshal(m["_hint"], &gotHint); err != nil {
		t.Fatalf("json.Unmarshal(_hint) returned error: %v", err)
	}
	if gotHint != CallContractItemHint.String() {
		t.Fatalf("unexpected JSON item hint: got %q, want %q", gotHint, CallContractItemHint)
	}

	var decoded CallContractItem
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(with hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decoded}, []CallContractItem{item})
	if !decoded.Hint().Equal(CallContractItemHint) {
		t.Fatalf("decoded item hint mismatch: %q", decoded.Hint())
	}

	var decodedByMethod CallContractItem
	if err := decodedByMethod.DecodeJSON(b, testCallContractJSONEncoder(t)); err != nil {
		t.Fatalf("DecodeJSON(with hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedByMethod}, []CallContractItem{item})
	if !decodedByMethod.Hint().Equal(CallContractItemHint) {
		t.Fatalf("DecodeJSON item hint mismatch: %q", decodedByMethod.Hint())
	}

	withoutHint := []byte(`{"function":"UpdateData","call_data":{"value":"next"}}`)
	var decodedWithoutHint CallContractItem
	if err := json.Unmarshal(withoutHint, &decodedWithoutHint); err != nil {
		t.Fatalf("json.Unmarshal(without hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedWithoutHint}, []CallContractItem{item})
	if !decodedWithoutHint.Hint().Equal(CallContractItemHint) {
		t.Fatalf("decoded item without hint should default to %q, got %q", CallContractItemHint, decodedWithoutHint.Hint())
	}

	var decodedWithoutHintByMethod CallContractItem
	if err := decodedWithoutHintByMethod.DecodeJSON(withoutHint, testCallContractJSONEncoder(t)); err != nil {
		t.Fatalf("DecodeJSON(without hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedWithoutHintByMethod}, []CallContractItem{item})
	if !decodedWithoutHintByMethod.Hint().Equal(CallContractItemHint) {
		t.Fatalf(
			"DecodeJSON item without hint should default to %q, got %q",
			CallContractItemHint,
			decodedWithoutHintByMethod.Hint(),
		)
	}
}

func TestCallContractItemBSONHintRoundTrip(t *testing.T) {
	item := NewCallContractItem("UpdateData", map[string]string{"value": "next"})
	b, err := bson.Marshal(item)
	if err != nil {
		t.Fatalf("bson.Marshal returned error: %v", err)
	}

	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatalf("bson.Unmarshal returned error: %v", err)
	}
	if gotHint := m["_hint"]; gotHint != CallContractItemHint.String() {
		t.Fatalf("unexpected BSON item hint: got %#v, want %q", gotHint, CallContractItemHint)
	}

	var decoded CallContractItem
	if err := bson.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("bson.Unmarshal(with hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decoded}, []CallContractItem{item})
	if !decoded.Hint().Equal(CallContractItemHint) {
		t.Fatalf("decoded item hint mismatch: %q", decoded.Hint())
	}

	var decodedByMethod CallContractItem
	if err := decodedByMethod.DecodeBSON(b, testCallContractBSONEncoder(t)); err != nil {
		t.Fatalf("DecodeBSON(with hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedByMethod}, []CallContractItem{item})
	if !decodedByMethod.Hint().Equal(CallContractItemHint) {
		t.Fatalf("DecodeBSON item hint mismatch: %q", decodedByMethod.Hint())
	}

	withoutHint, err := bson.Marshal(bson.M{
		"function":  "UpdateData",
		"call_data": bson.M{"value": "next"},
	})
	if err != nil {
		t.Fatalf("bson.Marshal(without hint) returned error: %v", err)
	}
	var decodedWithoutHint CallContractItem
	if err := bson.Unmarshal(withoutHint, &decodedWithoutHint); err != nil {
		t.Fatalf("bson.Unmarshal(without hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedWithoutHint}, []CallContractItem{item})
	if !decodedWithoutHint.Hint().Equal(CallContractItemHint) {
		t.Fatalf("decoded item without hint should default to %q, got %q", CallContractItemHint, decodedWithoutHint.Hint())
	}

	var decodedWithoutHintByMethod CallContractItem
	if err := decodedWithoutHintByMethod.DecodeBSON(withoutHint, testCallContractBSONEncoder(t)); err != nil {
		t.Fatalf("DecodeBSON(without hint) returned error: %v", err)
	}
	assertCallContractItems(t, []CallContractItem{decodedWithoutHintByMethod}, []CallContractItem{item})
	if !decodedWithoutHintByMethod.Hint().Equal(CallContractItemHint) {
		t.Fatalf(
			"DecodeBSON item without hint should default to %q, got %q",
			CallContractItemHint,
			decodedWithoutHintByMethod.Hint(),
		)
	}
}

func TestCallContractFactLegacyConstructorNormalizesToSingleItem(t *testing.T) {
	fact := testLegacyCallFact()

	assertCallContractItems(t, fact.Items(), []CallContractItem{
		NewCallContractItem("UpdateData", map[string]string{"value": "next"}),
	})

	callData := fact.CallData()
	if callData["function"] != "UpdateData" || callData["value"] != "next" {
		t.Fatalf("unexpected reconstructed legacy CallData: %#v", callData)
	}
}

func TestCallContractFactLegacyJSONDecodeNormalizesToSingleItem(t *testing.T) {
	fact := testLegacyCallFact()
	b, err := fact.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}

	var decoded CallContractFact
	if err := decoded.DecodeJSON(b, testCallContractJSONEncoder(t)); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}

	assertCallContractItems(t, decoded.Items(), []CallContractItem{
		NewCallContractItem("UpdateData", map[string]string{"value": "next"}),
	})
}

func TestCallContractFactItemsJSONDecodeAndMarshalPolicy(t *testing.T) {
	batch := testBatchCallFact()
	b, err := batch.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	assertJSONHasKey(t, b, "items")
	assertJSONLacksKey(t, b, "call_data")

	var decoded CallContractFact
	if err := decoded.DecodeJSON(b, testCallContractJSONEncoder(t)); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}
	assertCallContractItems(t, decoded.Items(), testBatchItems())

	singleItemsJSON := forceSingleItemsJSON(t, b)
	var singleDecoded CallContractFact
	if err := singleDecoded.DecodeJSON(singleItemsJSON, testCallContractJSONEncoder(t)); err != nil {
		t.Fatalf("DecodeJSON(single items) returned error: %v", err)
	}
	assertCallContractItems(t, singleDecoded.Items(), testBatchItems()[:1])
	out, err := singleDecoded.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(single decoded) returned error: %v", err)
	}
	assertJSONHasKey(t, out, "call_data")
	assertJSONLacksKey(t, out, "items")
}

func TestCallContractFactJSONRejectsCallDataAndItemsTogether(t *testing.T) {
	b := forceBothJSONShapes(t, testBatchCallFact())

	var decoded CallContractFact
	if err := decoded.DecodeJSON(b, testCallContractJSONEncoder(t)); err == nil {
		t.Fatal("expected DecodeJSON to reject call_data and items together")
	}
}

func TestCallContractFactLegacyJSONDecodeValidatesRawCallDataLimit(t *testing.T) {
	b := forceLegacyJSONCallData(t, testLegacyCallFact(), map[string]string{
		"function": strings.Repeat("f", 16*1024+1),
	})

	var decoded CallContractFact
	err := decoded.DecodeJSON(b, testCallContractJSONEncoder(t))
	if err == nil {
		t.Fatal("expected DecodeJSON to reject oversized legacy call_data")
	}
	if !strings.Contains(err.Error(), "value for key \"function\" exceeds max size") {
		t.Fatalf("expected raw legacy call_data limit error, got %v", err)
	}
}

func TestCallContractFactLegacyBSONDecodeNormalizesToSingleItem(t *testing.T) {
	fact := testLegacyCallFact()
	b, err := fact.MarshalBSON()
	if err != nil {
		t.Fatalf("MarshalBSON returned error: %v", err)
	}

	var decoded CallContractFact
	if err := decoded.DecodeBSON(b, testCallContractBSONEncoder(t)); err != nil {
		t.Fatalf("DecodeBSON returned error: %v", err)
	}

	assertCallContractItems(t, decoded.Items(), []CallContractItem{
		NewCallContractItem("UpdateData", map[string]string{"value": "next"}),
	})
}

func TestCallContractFactItemsBSONDecodeAndMarshalPolicy(t *testing.T) {
	batch := testBatchCallFact()
	b, err := batch.MarshalBSON()
	if err != nil {
		t.Fatalf("MarshalBSON returned error: %v", err)
	}
	assertBSONHasKey(t, b, "items")
	assertBSONLacksKey(t, b, "call_data")

	var decoded CallContractFact
	if err := decoded.DecodeBSON(b, testCallContractBSONEncoder(t)); err != nil {
		t.Fatalf("DecodeBSON returned error: %v", err)
	}
	assertCallContractItems(t, decoded.Items(), testBatchItems())

	singleItemsBSON := forceSingleItemsBSON(t, b)
	var singleDecoded CallContractFact
	if err := singleDecoded.DecodeBSON(singleItemsBSON, testCallContractBSONEncoder(t)); err != nil {
		t.Fatalf("DecodeBSON(single items) returned error: %v", err)
	}
	assertCallContractItems(t, singleDecoded.Items(), testBatchItems()[:1])
	out, err := singleDecoded.MarshalBSON()
	if err != nil {
		t.Fatalf("MarshalBSON(single decoded) returned error: %v", err)
	}
	assertBSONHasKey(t, out, "call_data")
	assertBSONLacksKey(t, out, "items")
}

func TestCallContractFactBSONRejectsCallDataAndItemsTogether(t *testing.T) {
	b := forceBothBSONShapes(t, testBatchCallFact())

	var decoded CallContractFact
	if err := decoded.DecodeBSON(b, testCallContractBSONEncoder(t)); err == nil {
		t.Fatal("expected DecodeBSON to reject call_data and items together")
	}
}

func TestCallContractFactLegacyBSONDecodeValidatesRawCallDataLimit(t *testing.T) {
	b := forceLegacyBSONCallData(t, testLegacyCallFact(), map[string]string{
		"function": strings.Repeat("f", 16*1024+1),
	})

	var decoded CallContractFact
	err := decoded.DecodeBSON(b, testCallContractBSONEncoder(t))
	if err == nil {
		t.Fatal("expected DecodeBSON to reject oversized legacy call_data")
	}
	if !strings.Contains(err.Error(), "value for key \"function\" exceeds max size") {
		t.Fatalf("expected raw legacy call_data limit error, got %v", err)
	}
}

func TestCallContractFactSingleCallHashCompatibility(t *testing.T) {
	legacy := testLegacyCallFact()
	legacyMap := map[string]string{"function": "UpdateData", "value": "next"}
	d, err := json.Marshal(legacyMap)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	want := valuehash.NewSHA256(util.ConcatBytesSlice(
		legacy.Token(),
		legacy.Sender().Bytes(),
		legacy.Contract().Bytes(),
		valuehash.NewSHA256(d).Bytes(),
		ctypes.CurrencyID("ABC").Bytes(),
	))
	items := NewCallContractFactWithItems(
		legacy.Token(),
		legacy.Sender(),
		legacy.Contract(),
		[]CallContractItem{NewCallContractItem("UpdateData", map[string]string{"value": "next"})},
		ctypes.CurrencyID("ABC"),
	)

	if legacy.Hash().String() != want.String() {
		t.Fatalf("legacy constructor hash no longer matches old algorithm\ngot:  %s\nwant: %s", legacy.Hash(), want)
	}
	if legacy.Hash().String() != items.Hash().String() {
		t.Fatalf("expected legacy and single-item hash to match\nlegacy: %s\nitems:  %s", legacy.Hash(), items.Hash())
	}
}

func TestCallContractFactBatchHashIsOrderSensitive(t *testing.T) {
	sender := base.NewStringAddress("batchsender0001")
	contract := base.NewStringAddress("batchcontract001")
	token := []byte("batch-token")
	ab := NewCallContractFactWithItems(token, sender, contract, testBatchItems(), ctypes.CurrencyID("ABC"))
	ba := NewCallContractFactWithItems(token, sender, contract, []CallContractItem{
		NewCallContractItem("UpdateData", map[string]string{"id": "a", "value": "two"}),
		NewCallContractItem("CreateData", map[string]string{"id": "a", "value": "one"}),
	}, ctypes.CurrencyID("ABC"))

	if ab.Hash().String() == ba.Hash().String() {
		t.Fatalf("expected item order to change hash: %s", ab.Hash())
	}
}

func TestCallContractFactBatchValidationLimits(t *testing.T) {
	tests := []struct {
		name      string
		items     []CallContractItem
		wantError string
	}{
		{
			name:      "empty items",
			items:     nil,
			wantError: "empty items",
		},
		{
			name:      "empty function",
			items:     []CallContractItem{NewCallContractItem("", map[string]string{})},
			wantError: "function is empty",
		},
		{
			name: "selector key in item call data",
			items: []CallContractItem{
				NewCallContractItem("UpdateData", map[string]string{"function": "Other"}),
			},
			wantError: "selector key",
		},
		{
			name:      "item count exceeded",
			items:     repeatedCallItems(MaxCallContractItems + 1),
			wantError: "max",
		},
		{
			name: "per item call data limit",
			items: []CallContractItem{
				NewCallContractItem("UpdateData", map[string]string{"value": strings.Repeat("v", 16*1024+1)}),
			},
			wantError: "value for key",
		},
		{
			name:      "aggregate limit",
			items:     aggregateLimitCallItems(),
			wantError: "max total function+call_data size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fact := NewCallContractFactWithItems(
				[]byte("token"),
				base.NewStringAddress("batchsender0002"),
				base.NewStringAddress("batchcontract002"),
				tt.items,
				ctypes.CurrencyID("ABC"),
			)
			err := fact.IsValid(nil)
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error containing %q, got %v", tt.wantError, err)
			}
		})
	}
}

func TestCallContractFactLegacySingleCallLimitUsesReconstructedMap(t *testing.T) {
	fact := NewCallContractFact(
		[]byte("legacy-limit-token"),
		base.NewStringAddress("legacysender0002"),
		base.NewStringAddress("legacycontract02"),
		map[string]string{
			"function": strings.Repeat("f", 16*1024+1),
		},
		ctypes.CurrencyID("ABC"),
	)

	err := fact.IsValid(nil)
	if err == nil {
		t.Fatal("expected reconstructed legacy call_data limit failure")
	}
	if !strings.Contains(err.Error(), "value for key \"function\" exceeds max size") {
		t.Fatalf("expected legacy function value limit error, got %v", err)
	}
}

func testLegacyCallFact() CallContractFact {
	return NewCallContractFact(
		[]byte("legacy-token"),
		base.NewStringAddress("legacysender0001"),
		base.NewStringAddress("legacycontract01"),
		map[string]string{
			"function": "UpdateData",
			"value":    "next",
		},
		ctypes.CurrencyID("ABC"),
	)
}

func testBatchCallFact() CallContractFact {
	return NewCallContractFactWithItems(
		[]byte("batch-token"),
		base.NewStringAddress("batchsender0001"),
		base.NewStringAddress("batchcontract001"),
		testBatchItems(),
		ctypes.CurrencyID("ABC"),
	)
}

func testBatchItems() []CallContractItem {
	return []CallContractItem{
		NewCallContractItem("CreateData", map[string]string{"id": "a", "value": "one"}),
		NewCallContractItem("UpdateData", map[string]string{"id": "a", "value": "two"}),
	}
}

func repeatedCallItems(n int) []CallContractItem {
	items := make([]CallContractItem, n)
	for i := range items {
		items[i] = NewCallContractItem("UpdateData", map[string]string{"value": "next"})
	}

	return items
}

func aggregateLimitCallItems() []CallContractItem {
	items := make([]CallContractItem, MaxCallContractItems)
	for i := range items {
		items[i] = NewCallContractItem("UpdateData", map[string]string{"value": strings.Repeat("v", 5000)})
	}

	return items
}

func assertCallContractItems(t *testing.T, got, want []CallContractItem) {
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
			t.Fatalf("item %d callData length mismatch: got %#v, want %#v", i, gotData, wantData)
		}
		for key, wantValue := range wantData {
			if gotData[key] != wantValue {
				t.Fatalf("item %d callData[%q] mismatch: got %q, want %q", i, key, gotData[key], wantValue)
			}
		}
		if _, found := gotData["function"]; found {
			t.Fatalf("item %d callData must not include function selector: %#v", i, gotData)
		}
	}
}

func testCallContractJSONEncoder(t *testing.T) encoder.Encoder {
	t.Helper()

	enc := jsonenc.NewEncoder()
	if err := enc.Add(encoder.DecodeDetail{Hint: base.StringAddressHint, Instance: base.StringAddress{}}); err != nil {
		t.Fatalf("Add(StringAddress) returned error: %v", err)
	}

	return enc
}

func testCallContractBSONEncoder(t *testing.T) *bsonenc.Encoder {
	t.Helper()

	enc := bsonenc.NewEncoder()
	if err := enc.Add(encoder.DecodeDetail{Hint: base.StringAddressHint, Instance: base.StringAddress{}}); err != nil {
		t.Fatalf("Add(StringAddress) returned error: %v", err)
	}

	return enc
}

func assertJSONHasKey(t *testing.T, b []byte, key string) {
	t.Helper()

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if _, found := m[key]; !found {
		t.Fatalf("expected JSON key %q in %s", key, b)
	}
}

func assertJSONLacksKey(t *testing.T, b []byte, key string) {
	t.Helper()

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if _, found := m[key]; found {
		t.Fatalf("did not expect JSON key %q in %s", key, b)
	}
}

func forceSingleItemsJSON(t *testing.T, b []byte) []byte {
	t.Helper()

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	items := m["items"].([]interface{})
	m["items"] = items[:1]
	delete(m, "call_data")

	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	return out
}

func forceBothJSONShapes(t *testing.T, fact CallContractFact) []byte {
	t.Helper()

	b, err := fact.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	m["call_data"] = map[string]string{"function": "UpdateData"}

	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	return out
}

func forceLegacyJSONCallData(t *testing.T, fact CallContractFact, callData map[string]string) []byte {
	t.Helper()

	b, err := fact.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON returned error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	delete(m, "items")
	m["call_data"] = callData
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	return out
}

func assertBSONHasKey(t *testing.T, b []byte, key string) {
	t.Helper()

	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatalf("bson.Unmarshal returned error: %v", err)
	}
	if _, found := m[key]; !found {
		t.Fatalf("expected BSON key %q in %#v", key, m)
	}
}

func assertBSONLacksKey(t *testing.T, b []byte, key string) {
	t.Helper()

	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatalf("bson.Unmarshal returned error: %v", err)
	}
	if _, found := m[key]; found {
		t.Fatalf("did not expect BSON key %q in %#v", key, m)
	}
}

func forceSingleItemsBSON(t *testing.T, b []byte) []byte {
	t.Helper()

	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatalf("bson.Unmarshal returned error: %v", err)
	}
	items := m["items"].(bson.A)
	m["items"] = items[:1]
	delete(m, "call_data")

	out, err := bson.Marshal(m)
	if err != nil {
		t.Fatalf("bson.Marshal returned error: %v", err)
	}

	return out
}

func forceBothBSONShapes(t *testing.T, fact CallContractFact) []byte {
	t.Helper()

	b, err := fact.MarshalBSON()
	if err != nil {
		t.Fatalf("MarshalBSON returned error: %v", err)
	}
	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatalf("bson.Unmarshal returned error: %v", err)
	}
	m["call_data"] = bson.M{"function": "UpdateData"}

	out, err := bson.Marshal(m)
	if err != nil {
		t.Fatalf("bson.Marshal returned error: %v", err)
	}

	return out
}

func forceLegacyBSONCallData(t *testing.T, fact CallContractFact, callData map[string]string) []byte {
	t.Helper()

	m := bson.M{
		"_hint":     fact.Hint().String(),
		"hash":      fact.Hash().String(),
		"token":     fact.Token(),
		"sender":    fact.Sender().String(),
		"contract":  fact.Contract().String(),
		"call_data": callData,
		"currency":  fact.currency,
	}
	out, err := bson.Marshal(m)
	if err != nil {
		t.Fatalf("bson.Marshal returned error: %v", err)
	}

	return out
}
