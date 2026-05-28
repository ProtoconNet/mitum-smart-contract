package types

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

// CurrentSchemaFormatVersion identifies the on-state metadata wrapper shape.
// It is separate from the typed schema ruleset version and from any runtime
// snapshot codec version.
const CurrentSchemaFormatVersion = "contract-schema-format-v1"

type PersistedContractSchema struct {
	SchemaFormatVersion  string         `json:"schema_format_version" bson:"schema_format_version"`
	SchemaRulesetVersion string         `json:"schema_ruleset_version" bson:"schema_ruleset_version"`
	SourceHash           string         `json:"source_hash" bson:"source_hash"`
	Schema               ContractSchema `json:"schema" bson:"schema"`
}

type ContractSchema struct {
	PackageName       string                    `json:"package_name" bson:"package_name"`
	Mode              string                    `json:"mode" bson:"mode"`
	Types             TypeRegistry              `json:"types" bson:"types"`
	PersistentGlobals []PersistentBindingSchema `json:"persistent_globals" bson:"persistent_globals"`
	Functions         []FunctionSchema          `json:"functions" bson:"functions"`
}

type TypeRegistry struct {
	Structs map[string]TypeRef `json:"structs" bson:"structs"`
}

type TypeRef struct {
	Kind   string        `json:"kind" bson:"kind"`
	Raw    string        `json:"raw,omitempty" bson:"raw,omitempty"`
	Scalar string        `json:"scalar,omitempty" bson:"scalar,omitempty"`
	Name   string        `json:"name,omitempty" bson:"name,omitempty"`
	Key    *TypeRef      `json:"key,omitempty" bson:"key,omitempty"`
	Elem   *TypeRef      `json:"elem,omitempty" bson:"elem,omitempty"`
	Fields []StructField `json:"fields,omitempty" bson:"fields,omitempty"`
}

type StructField struct {
	Name string  `json:"name" bson:"name"`
	Type TypeRef `json:"type" bson:"type"`
}

type PersistentBindingSchema struct {
	Name            string  `json:"name" bson:"name"`
	Type            TypeRef `json:"type" bson:"type"`
	HasExplicitType bool    `json:"has_explicit_type" bson:"has_explicit_type"`
}

type FunctionSchema struct {
	Name     string         `json:"name" bson:"name"`
	Exported bool           `json:"exported" bson:"exported"`
	Params   []ParamSchema  `json:"params" bson:"params"`
	Results  []ResultSchema `json:"results" bson:"results"`
}

type ParamSchema struct {
	Name string  `json:"name" bson:"name"`
	Type TypeRef `json:"type" bson:"type"`
}

type ResultSchema struct {
	Type TypeRef `json:"type" bson:"type"`
}

func ContractSourceHash(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

func (s PersistedContractSchema) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid PersistedContractSchema")

	if s.SchemaFormatVersion == "" {
		return e.Wrap(errors.Errorf("empty schema format version"))
	}
	if s.SchemaRulesetVersion == "" {
		return e.Wrap(errors.Errorf("empty schema ruleset version"))
	}
	if s.SourceHash == "" {
		return e.Wrap(errors.Errorf("empty source hash"))
	}
	if err := s.Schema.IsValid(nil); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (s ContractSchema) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid ContractSchema")

	if s.PackageName == "" {
		return e.Wrap(errors.Errorf("empty package name"))
	}
	if s.Mode == "" {
		return e.Wrap(errors.Errorf("empty mode"))
	}

	return nil
}

func (s PersistedContractSchema) Bytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(s.SchemaFormatVersion),
		[]byte(s.SchemaRulesetVersion),
		[]byte(s.SourceHash),
		s.Schema.Bytes(),
	)
}

func (s ContractSchema) Bytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(s.PackageName),
		[]byte(s.Mode),
		s.Types.Bytes(),
		persistentBindingSchemasBytes(s.PersistentGlobals),
		functionSchemasBytes(s.Functions),
	)
}

func (r TypeRegistry) Bytes() []byte {
	if len(r.Structs) == 0 {
		return nil
	}

	keys := make([]string, 0, len(r.Structs))
	for key := range r.Structs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	bs := make([][]byte, 0, len(keys)*2)
	for _, key := range keys {
		bs = append(bs, []byte(key), r.Structs[key].Bytes())
	}

	return util.ConcatBytesSlice(bs...)
}

func (t TypeRef) Bytes() []byte {
	var keyBytes []byte
	if t.Key != nil {
		keyBytes = t.Key.Bytes()
	}

	var elemBytes []byte
	if t.Elem != nil {
		elemBytes = t.Elem.Bytes()
	}

	return util.ConcatBytesSlice(
		[]byte(t.Kind),
		[]byte(t.Raw),
		[]byte(t.Scalar),
		[]byte(t.Name),
		keyBytes,
		elemBytes,
		structFieldsBytes(t.Fields),
	)
}

func persistentBindingSchemasBytes(in []PersistentBindingSchema) []byte {
	bs := make([][]byte, 0, len(in))
	for _, binding := range in {
		bs = append(bs, util.ConcatBytesSlice(
			[]byte(binding.Name),
			binding.Type.Bytes(),
			[]byte(strconv.FormatBool(binding.HasExplicitType)),
		))
	}

	return util.ConcatBytesSlice(bs...)
}

func functionSchemasBytes(in []FunctionSchema) []byte {
	bs := make([][]byte, 0, len(in))
	for _, fn := range in {
		bs = append(bs, util.ConcatBytesSlice(
			[]byte(fn.Name),
			[]byte(strconv.FormatBool(fn.Exported)),
			paramSchemasBytes(fn.Params),
			resultSchemasBytes(fn.Results),
		))
	}

	return util.ConcatBytesSlice(bs...)
}

func paramSchemasBytes(in []ParamSchema) []byte {
	bs := make([][]byte, 0, len(in))
	for _, param := range in {
		bs = append(bs, util.ConcatBytesSlice([]byte(param.Name), param.Type.Bytes()))
	}

	return util.ConcatBytesSlice(bs...)
}

func resultSchemasBytes(in []ResultSchema) []byte {
	bs := make([][]byte, 0, len(in))
	for _, result := range in {
		bs = append(bs, result.Type.Bytes())
	}

	return util.ConcatBytesSlice(bs...)
}

func structFieldsBytes(in []StructField) []byte {
	bs := make([][]byte, 0, len(in))
	for _, field := range in {
		bs = append(bs, util.ConcatBytesSlice([]byte(field.Name), field.Type.Bytes()))
	}

	return util.ConcatBytesSlice(bs...)
}
