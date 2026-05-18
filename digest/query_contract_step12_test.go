package digest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	digestmongo "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const typedDigestQueryContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
	Flags map[string]bool
	Aliases []string
}

type User struct {
	Balance int64
	Meta Meta
}

type Config struct {
	Owner string
	FeatureFlags map[string]bool
	Users map[string]User
	Aliases []string
	Watchers []User
}

type UserSelector struct {
	Name string
	RequireActive bool
}

var config Config

func Initialize(ctx chain.ContractContext) error {
	config.Owner = ctx.GetSender()
	config.FeatureFlags = map[string]bool{"alpha": true, "beta": false}
	config.Users = map[string]User{
		"alice": User{Balance: 10, Meta: Meta{Limit: 100, Flags: map[string]bool{"active": true}, Aliases: []string{"a1"}}},
		"bob": User{Balance: 20, Meta: Meta{Limit: 200, Flags: map[string]bool{"active": false}, Aliases: []string{"b1"}}},
	}
	config.Aliases = []string{"root", "child"}
	config.Watchers = []User{
		User{Balance: 30, Meta: Meta{Limit: 300, Flags: map[string]bool{"active": true}, Aliases: []string{"w1"}}},
		User{Balance: 40, Meta: Meta{Limit: 400, Flags: map[string]bool{"active": false}, Aliases: []string{"w2"}}},
	}
	return nil
}

func GetOwner(ctx chain.ContractContext) string { return config.Owner }
func GetConfig(ctx chain.ContractContext) Config { return config }
func GetFeatureFlags(ctx chain.ContractContext) map[string]bool { return config.FeatureFlags }
func GetUsers(ctx chain.ContractContext) map[string]User { return config.Users }
func GetAliases(ctx chain.ContractContext) []string { return config.Aliases }
func GetWatchers(ctx chain.ContractContext) []User { return config.Watchers }

func GetUser(ctx chain.ContractContext, name string) (User, bool) {
	user, found := config.Users[name]
	return user, found
}

func GetSelectedUser(ctx chain.ContractContext, selector UserSelector) (User, bool) {
	user, found := config.Users[selector.Name]
	if !found {
		return User{}, false
	}
	if selector.RequireActive && !user.Meta.Flags["active"] {
		return User{}, false
	}
	return user, true
}

func EchoFlags(ctx chain.ContractContext, next map[string]bool) map[string]bool {
	return next
}

func EchoUsers(ctx chain.ContractContext, next map[string]User) map[string]User {
	return next
}

func EchoAliases(ctx chain.ContractContext, next []string) []string {
	return next
}
`

func TestContractQueryEndpointScalarResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(500),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, headers := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetOwner",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if got := headers.Get("Content-Type"); !strings.Contains(got, "application/hal+json") {
		t.Fatalf("unexpected content type: %s", got)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "engine", "gno-snapshot-v1")
	assertEmbeddedField(t, resp, "read_only", true)
	assertEmbeddedField(t, resp, "result", "senderd0001sas")
	assertHasHALLink(t, resp, "design")
	assertHasHALLink(t, resp, "block")

	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointStructResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(510),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetConfig",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{
		"Owner": "senderd0001sas",
		"FeatureFlags": map[string]interface{}{
			"alpha": true,
			"beta":  false,
		},
		"Users": map[string]interface{}{
			"alice": map[string]interface{}{
				"Balance": float64(10),
				"Meta": map[string]interface{}{
					"Limit":   float64(100),
					"Flags":   map[string]interface{}{"active": true},
					"Aliases": []interface{}{"a1"},
				},
			},
			"bob": map[string]interface{}{
				"Balance": float64(20),
				"Meta": map[string]interface{}{
					"Limit":   float64(200),
					"Flags":   map[string]interface{}{"active": false},
					"Aliases": []interface{}{"b1"},
				},
			},
		},
		"Aliases": []interface{}{"root", "child"},
		"Watchers": []interface{}{
			map[string]interface{}{
				"Balance": float64(30),
				"Meta": map[string]interface{}{
					"Limit":   float64(300),
					"Flags":   map[string]interface{}{"active": true},
					"Aliases": []interface{}{"w1"},
				},
			},
			map[string]interface{}{
				"Balance": float64(40),
				"Meta": map[string]interface{}{
					"Limit":   float64(400),
					"Flags":   map[string]interface{}{"active": false},
					"Aliases": []interface{}{"w2"},
				},
			},
		},
	})

	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointMapScalarResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(520),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetFeatureFlags",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{"alpha": true, "beta": false})
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointMapStructResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(530),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetUsers",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{
		"alice": map[string]interface{}{
			"Balance": float64(10),
			"Meta": map[string]interface{}{
				"Limit":   float64(100),
				"Flags":   map[string]interface{}{"active": true},
				"Aliases": []interface{}{"a1"},
			},
		},
		"bob": map[string]interface{}{
			"Balance": float64(20),
			"Meta": map[string]interface{}{
				"Limit":   float64(200),
				"Flags":   map[string]interface{}{"active": false},
				"Aliases": []interface{}{"b1"},
			},
		},
	})
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointSliceScalarResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(540),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetAliases",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", []interface{}{"root", "child"})
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointSliceStructResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(550),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetWatchers",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", []interface{}{
		map[string]interface{}{
			"Balance": float64(30),
			"Meta": map[string]interface{}{
				"Limit":   float64(300),
				"Flags":   map[string]interface{}{"active": true},
				"Aliases": []interface{}{"w1"},
			},
		},
		map[string]interface{}{
			"Balance": float64(40),
			"Meta": map[string]interface{}{
				"Limit":   float64(400),
				"Flags":   map[string]interface{}{"active": false},
				"Aliases": []interface{}{"w2"},
			},
		},
	})
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointOptionalResult(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(560),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetUser",
		"name":     "alice",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "ok", true)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{
		"Balance": float64(10),
		"Meta": map[string]interface{}{
			"Limit":   float64(100),
			"Flags":   map[string]interface{}{"active": true},
			"Aliases": []interface{}{"a1"},
		},
	})

	status, body, _ = performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetUser",
		"name":     "nobody",
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp = decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "ok", false)
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointStructArg(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(570),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "GetSelectedUser",
		"selector": `{"name":"alice","requireactive":true}`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}

	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "ok", true)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{
		"Balance": float64(10),
		"Meta": map[string]interface{}{
			"Limit":   float64(100),
			"Flags":   map[string]interface{}{"active": true},
			"Aliases": []interface{}{"a1"},
		},
	})
	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointMapScalarArgAndNilEmptyPolicies(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(580),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoFlags",
		"next":     `{"beta":false,"alpha":true}`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{"alpha": true, "beta": false})

	status, body, _ = performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoFlags",
		"next":     `null`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp = decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", nil)

	status, body, _ = performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoFlags",
		"next":     `{}`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp = decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", map[string]interface{}{})

	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointSliceScalarArgAndNilEmptyPolicies(t *testing.T) {
	hd, contract, snapshotBefore := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(590),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoAliases",
		"next":     `["a1","a2"]`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp := decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", []interface{}{"a1", "a2"})

	status, body, _ = performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoAliases",
		"next":     `null`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp = decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", nil)

	status, body, _ = performContractQueryRequest(t, hd, contract, map[string]string{
		"function": "EchoAliases",
		"next":     `[]`,
	})
	if status != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	resp = decodeHALResponse(t, body)
	assertEmbeddedField(t, resp, "result", []interface{}{})

	assertDigestSnapshotStateUnchanged(t, hd.database, contract, snapshotBefore)
}

func TestContractQueryEndpointMalformedJSONBodyRejected(t *testing.T) {
	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(600),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, headers := performRawContractQueryRequest(t, hd, contract, `{"function":"EchoAliases","next":"["a1",}"}`)
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if got := headers.Get("Content-Type"); !strings.Contains(got, "application/problem+json") {
		t.Fatalf("unexpected problem content type: %s", got)
	}
}

func TestContractQueryEndpointMissingFunctionRejected(t *testing.T) {
	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{
			Mode:     cruntime.InvocationModeRegister,
			Height:   base.Height(610),
			Function: "Initialize",
			CallData: map[string]string{},
		},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{
		"next": `[]`,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
}

func newTypedQueryTestHandlers(
	t *testing.T,
	source string,
	requests []cruntime.ExecuteRequest,
) (*Handlers, string, []byte) {
	t.Helper()

	encsPtr, enc := newDigestTestEncoders(t)
	contract := base.NewStringAddress("contractd0001")
	sender := base.NewStringAddress("senderd0001")

	states := map[string]base.State{}
	getStateFunc := func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}

	engine := cruntime.NewGnoEngine()

	for _, req := range requests {
		req.Contract = contract
		req.Sender = sender
		req.ContractCode = source

		res, err := engine.ExecuteContract(*encsPtr, getStateFunc, req)
		if err != nil {
			t.Fatalf("ExecuteContract(%s) returned error: %v", req.Function, err)
		}

		for _, merge := range res.StateMerges {
			states[merge.Key()] = common.NewBaseState(req.Height, merge.Key(), merge.Value(), nil, nil)
		}
	}

	states[pstate.DesignStateKey(contract)] = common.NewBaseState(
		requests[0].Height,
		pstate.DesignStateKey(contract),
		pstate.NewDesignStateValue(ptypes.NewDesign(source)),
		nil,
		nil,
	)

	snapshotBefore := snapshotBytesFromStates(t, states, contract)
	return newDigestHandlersForStates(t, encsPtr, enc, states), contract.String(), snapshotBefore
}

func newDigestHandlersForStates(
	t *testing.T,
	encs *encoder.Encoders,
	enc encoder.Encoder,
	states map[string]base.State,
) *Handlers {
	t.Helper()

	mdb, err := digestmongo.NewDatabase(nil, encs, enc)
	if err != nil {
		t.Fatalf("digestmongo.NewDatabase returned error: %v", err)
	}

	center := &isaacdatabase.Center{
		Logging: logging.NewLogging(nil).SetLogger(zerolog.Nop()),
	}

	db, err := NewDatabase(center, mdb)
	if err != nil {
		t.Fatalf("digest.NewDatabase returned error: %v", err)
	}
	db.stateGetter = func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}

	router := mux.NewRouter()
	ctx := context.WithValue(context.Background(), launch.LoggingContextKey, logging.NewLogging(nil).SetLogger(zerolog.Nop()))
	hd := NewHandlers(ctx, base.NetworkID("testnet"), encs, enc, db, DummyCache{}, router, nil)
	if hd == nil {
		t.Fatalf("NewHandlers returned nil")
	}
	if err := hd.Initialize(); err != nil {
		t.Fatalf("Handlers.Initialize returned error: %v", err)
	}

	return hd
}

func newDigestTestEncoders(t *testing.T) (*encoder.Encoders, encoder.Encoder) {
	t.Helper()

	enc := jsonenc.NewEncoder()
	encs := encoder.NewEncoders(enc, enc)
	if err := encs.AddDetail(encoder.DecodeDetail{
		Hint:     base.StringAddressHint,
		Instance: base.StringAddress{},
	}); err != nil {
		t.Fatalf("AddDetail(StringAddress) returned error: %v", err)
	}

	return encs, enc
}

func performContractQueryRequest(t *testing.T, hd *Handlers, contract string, body map[string]string) (int, string, http.Header) {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal(body) returned error: %v", err)
	}

	return performRawContractQueryRequest(t, hd, contract, string(b))
}

func performRawContractQueryRequest(t *testing.T, hd *Handlers, contract string, body string) (int, string, http.Header) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/contract/"+contract+"/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	hd.Router().ServeHTTP(rec, req)
	return rec.Code, rec.Body.String(), rec.Header()
}

func decodeHALResponse(t *testing.T, body string) map[string]interface{} {
	t.Helper()

	var v map[string]interface{}
	dec := json.NewDecoder(bytes.NewBufferString(body))
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode HAL response returned error: %v\nbody=%s", err, body)
	}
	return v
}

func assertEmbeddedField(t *testing.T, resp map[string]interface{}, key string, want interface{}) {
	t.Helper()

	embedded, ok := resp["_embedded"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing _embedded in response: %#v", resp)
	}
	got, found := embedded[key]
	if !found {
		t.Fatalf("missing _embedded.%s in response: %#v", key, embedded)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected _embedded.%s\nwant: %#v\ngot:  %#v", key, want, got)
	}
}

func assertHasHALLink(t *testing.T, resp map[string]interface{}, rel string) {
	t.Helper()

	links, ok := resp["_links"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing _links in response: %#v", resp)
	}
	link, found := links[rel]
	if !found {
		t.Fatalf("missing _links.%s in response: %#v", rel, links)
	}
	m, ok := link.(map[string]interface{})
	if !ok {
		t.Fatalf("invalid _links.%s value: %#v", rel, link)
	}
	href, ok := m["href"].(string)
	if !ok || strings.TrimSpace(href) == "" {
		t.Fatalf("invalid _links.%s.href: %#v", rel, m["href"])
	}
}

func snapshotBytesFromStates(t *testing.T, states map[string]base.State, contract base.Address) []byte {
	t.Helper()

	sv, err := pstate.GetSnapshotFromState(states[pstate.SnapshotStateKey(contract)])
	if err != nil {
		t.Fatalf("GetSnapshotFromState returned error: %v", err)
	}
	return append([]byte(nil), sv.Snapshot...)
}

func assertDigestSnapshotStateUnchanged(t *testing.T, db *Database, contract string, before []byte) {
	t.Helper()

	_, sv, _, found, err := ContractSnapshotFromChainState(db, contract)
	if err != nil {
		t.Fatalf("ContractSnapshotFromChainState returned error: %v", err)
	}
	if !found {
		t.Fatalf("snapshot state not found for %s", contract)
	}

	if !reflect.DeepEqual(before, sv.Snapshot) {
		t.Fatalf("expected endpoint query to leave snapshot bytes unchanged\nbefore: %s\nafter:  %s", before, sv.Snapshot)
	}
}
