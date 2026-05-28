package state_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum-smart-contract/types"
	"github.com/ProtoconNet/mitum2/base"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

const designStateSchemaTestSource = `package contract
import "mitum/chain"

var value string

func Initialize(ctx chain.WriteContext, seed string) error {
	value = seed
	return nil
}

func GetValue(ctx chain.QueryContext) string { return value }
`

func TestDesignStateValueDecodeOldJSONWithoutSchema(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	b, err := json.Marshal(oldDesignStateValueJSON{
		BaseHinter: hint.NewBaseHinter(state.DesignStateValueHint),
		Design:     design,
	})
	if err != nil {
		t.Fatalf("marshal old json: %v", err)
	}
	if strings.Contains(string(b), `"schema"`) {
		t.Fatalf("old json fixture unexpectedly includes schema: %s", b)
	}

	var decoded state.DesignStateValue
	if err := decoded.DecodeJSON(b, jsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode old json: %v", err)
	}
	if decoded.Schema != nil {
		t.Fatalf("expected nil schema for old json, got %#v", decoded.Schema)
	}
	if decoded.Design.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("contract source changed after old json decode")
	}
}

func TestDesignStateValueDecodeOldBSONWithoutSchema(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	b, err := bsonenc.Marshal(bson.M{
		"_hint":  state.DesignStateValueHint.String(),
		"design": design,
	})
	if err != nil {
		t.Fatalf("marshal old bson: %v", err)
	}

	var decoded state.DesignStateValue
	if err := decoded.DecodeBSON(b, bsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode old bson: %v", err)
	}
	if decoded.Schema != nil {
		t.Fatalf("expected nil schema for old bson, got %#v", decoded.Schema)
	}
	if decoded.Design.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("contract source changed after old bson decode")
	}
}

func TestDesignStateValueJSONRoundTripWithPersistedSchema(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	schema := samplePersistedContractSchema(designStateSchemaTestSource)
	stateValue := state.NewDesignStateValueWithSchema(design, schema)

	b, err := json.Marshal(stateValue)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if !strings.Contains(string(b), `"schema"`) {
		t.Fatalf("json with persisted schema omitted schema metadata: %s", b)
	}

	var decoded state.DesignStateValue
	if err := decoded.DecodeJSON(b, jsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode json: %v", err)
	}

	assertPersistedSchemaRoundTrip(t, decoded.Schema, schema)
	if decoded.Design.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("contract source changed after json roundtrip")
	}
}

func TestDesignStateValueBSONRoundTripWithPersistedSchema(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	schema := samplePersistedContractSchema(designStateSchemaTestSource)
	stateValue := state.NewDesignStateValueWithSchema(design, schema)

	b, err := stateValue.MarshalBSON()
	if err != nil {
		t.Fatalf("marshal bson: %v", err)
	}

	var decoded state.DesignStateValue
	if err := decoded.DecodeBSON(b, bsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode bson: %v", err)
	}

	assertPersistedSchemaRoundTrip(t, decoded.Schema, schema)
	if decoded.Design.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("contract source changed after bson roundtrip")
	}
}

func TestDesignStateValueNilSchemaRoundTrip(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	stateValue := state.NewDesignStateValue(design)

	jsonBytes, err := json.Marshal(stateValue)
	if err != nil {
		t.Fatalf("marshal nil-schema json: %v", err)
	}
	if strings.Contains(string(jsonBytes), `"schema"`) {
		t.Fatalf("nil-schema json should omit schema field: %s", jsonBytes)
	}
	var decodedJSON state.DesignStateValue
	if err := decodedJSON.DecodeJSON(jsonBytes, jsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode nil-schema json: %v", err)
	}
	if decodedJSON.Schema != nil {
		t.Fatalf("expected nil schema after json roundtrip, got %#v", decodedJSON.Schema)
	}

	bsonBytes, err := stateValue.MarshalBSON()
	if err != nil {
		t.Fatalf("marshal nil-schema bson: %v", err)
	}
	var decodedBSON state.DesignStateValue
	if err := decodedBSON.DecodeBSON(bsonBytes, bsonenc.NewEncoder()); err != nil {
		t.Fatalf("decode nil-schema bson: %v", err)
	}
	if decodedBSON.Schema != nil {
		t.Fatalf("expected nil schema after bson roundtrip, got %#v", decodedBSON.Schema)
	}
}

func TestPersistedContractSchemaVersionConstantsAreSeparate(t *testing.T) {
	if types.CurrentSchemaFormatVersion != "contract-schema-format-v1" {
		t.Fatalf("unexpected schema format version: %q", types.CurrentSchemaFormatVersion)
	}
	if runtime.CurrentSchemaRulesetVersion != "typed-gno-ruleset-v1" {
		t.Fatalf("unexpected schema ruleset version: %q", runtime.CurrentSchemaRulesetVersion)
	}
	if types.CurrentSchemaFormatVersion == runtime.CurrentSchemaRulesetVersion {
		t.Fatalf("schema format version must be distinct from schema ruleset version")
	}
	if types.CurrentSchemaFormatVersion == string(runtime.SchemaModeTypedArgs) {
		t.Fatalf("schema format version must be distinct from schema mode")
	}
}

func TestPersistedContractSchemaValidation(t *testing.T) {
	valid := samplePersistedContractSchema(designStateSchemaTestSource)
	if err := valid.IsValid(nil); err != nil {
		t.Fatalf("valid persisted schema failed validation: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*types.PersistedContractSchema)
	}{
		{
			name: "empty format version",
			mutate: func(schema *types.PersistedContractSchema) {
				schema.SchemaFormatVersion = ""
			},
		},
		{
			name: "empty ruleset version",
			mutate: func(schema *types.PersistedContractSchema) {
				schema.SchemaRulesetVersion = ""
			},
		},
		{
			name: "empty source hash",
			mutate: func(schema *types.PersistedContractSchema) {
				schema.SourceHash = ""
			},
		},
		{
			name: "empty schema package",
			mutate: func(schema *types.PersistedContractSchema) {
				schema.Schema.PackageName = ""
			},
		},
		{
			name: "empty schema mode",
			mutate: func(schema *types.PersistedContractSchema) {
				schema.Schema.Mode = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := samplePersistedContractSchema(designStateSchemaTestSource)
			tt.mutate(schema)
			if err := schema.IsValid(nil); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestContractSourceHashDeterministic(t *testing.T) {
	hash1 := types.ContractSourceHash(designStateSchemaTestSource)
	hash2 := types.ContractSourceHash(designStateSchemaTestSource)
	hash3 := types.ContractSourceHash(designStateSchemaTestSource + "\n")

	if hash1 != hash2 {
		t.Fatalf("same source produced different hashes: %q != %q", hash1, hash2)
	}
	if hash1 == hash3 {
		t.Fatalf("different source produced same hash: %q", hash1)
	}
	if len(hash1) != 64 {
		t.Fatalf("expected sha256 hex hash length 64, got %d", len(hash1))
	}
}

func TestDesignStateValueSchemaDoesNotReplaceRawSource(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	schema := samplePersistedContractSchema(designStateSchemaTestSource)
	stateValue := state.NewDesignStateValueWithSchema(design, schema)

	if stateValue.Design.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("state value did not preserve raw contract source")
	}

	st := common.NewBaseState(base.Height(100), "contract:test:design", stateValue, nil, nil)
	got, err := state.GetDesignFromState(st)
	if err != nil {
		t.Fatalf("get design from state: %v", err)
	}
	if got.ContractCode() != designStateSchemaTestSource {
		t.Fatalf("design state no longer returns raw contract source")
	}
}

func TestDesignStateValueNilSchemaHashBytesPreserveOldBehavior(t *testing.T) {
	design := types.NewDesign(designStateSchemaTestSource)
	nilSchemaValue := state.NewDesignStateValue(design)
	if !bytes.Equal(nilSchemaValue.HashBytes(), design.Bytes()) {
		t.Fatalf("nil schema hash bytes should match legacy design bytes")
	}

	withSchemaValue := state.NewDesignStateValueWithSchema(
		design,
		samplePersistedContractSchema(designStateSchemaTestSource),
	)
	if bytes.Equal(withSchemaValue.HashBytes(), design.Bytes()) {
		t.Fatalf("schema-bearing state should include schema metadata in hash bytes")
	}
}

type oldDesignStateValueJSON struct {
	hint.BaseHinter
	Design types.Design `json:"design"`
}

func samplePersistedContractSchema(source string) *types.PersistedContractSchema {
	stringType := scalarType("string")
	recordType := types.TypeRef{
		Kind: string(runtime.TypeStruct),
		Raw:  "Record",
		Name: "Record",
		Fields: []types.StructField{
			{
				Name: "Value",
				Type: stringType,
			},
		},
	}

	return &types.PersistedContractSchema{
		SchemaFormatVersion:  types.CurrentSchemaFormatVersion,
		SchemaRulesetVersion: runtime.CurrentSchemaRulesetVersion,
		SourceHash:           types.ContractSourceHash(source),
		Schema: types.ContractSchema{
			PackageName: "contract",
			Mode:        string(runtime.SchemaModeTypedArgs),
			Types: types.TypeRegistry{
				Structs: map[string]types.TypeRef{
					"Record": recordType,
				},
			},
			PersistentGlobals: []types.PersistentBindingSchema{
				{
					Name:            "value",
					Type:            stringType,
					HasExplicitType: true,
				},
				{
					Name:            "records",
					Type:            mapType(scalarType("string"), namedType("Record")),
					HasExplicitType: true,
				},
			},
			Functions: []types.FunctionSchema{
				{
					Name:     "Initialize",
					Exported: true,
					Params: []types.ParamSchema{
						{Name: "seed", Type: stringType},
					},
					Results: []types.ResultSchema{
						{Type: opaqueType("error")},
					},
				},
				{
					Name:     "GetValue",
					Exported: true,
					Results: []types.ResultSchema{
						{Type: stringType},
					},
				},
			},
		},
	}
}

func scalarType(name string) types.TypeRef {
	return types.TypeRef{
		Kind:   string(runtime.TypeScalar),
		Raw:    name,
		Scalar: name,
	}
}

func namedType(name string) types.TypeRef {
	return types.TypeRef{
		Kind: string(runtime.TypeNamed),
		Raw:  name,
		Name: name,
	}
}

func mapType(key, elem types.TypeRef) types.TypeRef {
	return types.TypeRef{
		Kind: string(runtime.TypeMap),
		Raw:  "map[" + key.Raw + "]" + elem.Raw,
		Key:  &key,
		Elem: &elem,
	}
}

func opaqueType(name string) types.TypeRef {
	return types.TypeRef{
		Kind: string(runtime.TypeOpaque),
		Raw:  name,
		Name: name,
	}
}

func assertPersistedSchemaRoundTrip(
	t *testing.T,
	got *types.PersistedContractSchema,
	want *types.PersistedContractSchema,
) {
	t.Helper()

	if got == nil {
		t.Fatalf("expected persisted schema, got nil")
	}
	if got.SchemaFormatVersion != types.CurrentSchemaFormatVersion {
		t.Fatalf("schema format version mismatch: %q", got.SchemaFormatVersion)
	}
	if got.SchemaRulesetVersion != runtime.CurrentSchemaRulesetVersion {
		t.Fatalf("schema ruleset version mismatch: %q", got.SchemaRulesetVersion)
	}
	if got.SourceHash != want.SourceHash {
		t.Fatalf("source hash mismatch: %q != %q", got.SourceHash, want.SourceHash)
	}
	if got.Schema.PackageName != "contract" {
		t.Fatalf("package mismatch: %q", got.Schema.PackageName)
	}
	if got.Schema.Mode != string(runtime.SchemaModeTypedArgs) {
		t.Fatalf("schema mode mismatch: %q", got.Schema.Mode)
	}
	if !reflect.DeepEqual(got.Schema.Types, want.Schema.Types) {
		t.Fatalf("type registry mismatch:\ngot:  %#v\nwant: %#v", got.Schema.Types, want.Schema.Types)
	}
	if !reflect.DeepEqual(got.Schema.PersistentGlobals, want.Schema.PersistentGlobals) {
		t.Fatalf("persistent globals mismatch:\ngot:  %#v\nwant: %#v", got.Schema.PersistentGlobals, want.Schema.PersistentGlobals)
	}
	if !reflect.DeepEqual(got.Schema.Functions, want.Schema.Functions) {
		t.Fatalf("functions mismatch:\ngot:  %#v\nwant: %#v", got.Schema.Functions, want.Schema.Functions)
	}
}
