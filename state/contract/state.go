package contract

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	types "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	DesignStateValueHint   = hint.MustNewHint("mitum-contract-design-state-value-v0.0.1")
	ContractStateKeyPrefix = "contract"
	DesignStateKeySuffix   = "design"
)

func ContractStateKey(addr base.Address) string {
	return fmt.Sprintf("%s:%s", ContractStateKeyPrefix, addr.String())
}

type DesignStateValue struct {
	hint.BaseHinter
	Design types.Design
	Schema *types.PersistedContractSchema
}

func NewDesignStateValue(design types.Design) DesignStateValue {
	return NewDesignStateValueWithSchema(design, nil)
}

func NewDesignStateValueWithSchema(
	design types.Design,
	schema *types.PersistedContractSchema,
) DesignStateValue {
	return DesignStateValue{
		BaseHinter: hint.NewBaseHinter(DesignStateValueHint),
		Design:     design,
		Schema:     schema,
	}
}

func (sv DesignStateValue) Hint() hint.Hint {
	return sv.BaseHinter.Hint()
}

func (sv DesignStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid DesignStateValue")

	if err := sv.BaseHinter.IsValid(DesignStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := sv.Design.IsValid(nil); err != nil {
		return e.Wrap(err)
	}
	if sv.Schema != nil {
		if err := sv.Schema.IsValid(nil); err != nil {
			return e.Wrap(err)
		}
	}

	return nil
}

func (sv DesignStateValue) HashBytes() []byte {
	if sv.Schema != nil {
		return util.ConcatBytesSlice(sv.Design.Bytes(), sv.Schema.Bytes())
	}

	return sv.Design.Bytes()
}

func GetDesignFromState(st base.State) (types.Design, error) {
	d, err := GetDesignStateValueFromState(st)
	if err != nil {
		return types.Design{}, err
	}

	return d.Design, nil
}

func GetDesignStateValueFromState(st base.State) (DesignStateValue, error) {
	v := st.Value()
	if v == nil {
		return DesignStateValue{}, errors.Errorf("state value is nil")
	}

	d, ok := v.(DesignStateValue)
	if !ok {
		return DesignStateValue{}, errors.Errorf("expected DesignStateValue but %T", v)
	}

	return d, nil
}

func IsDesignStateKey(key string) bool {
	return strings.HasPrefix(key, ContractStateKeyPrefix) && strings.HasSuffix(key, DesignStateKeySuffix)
}

func DesignStateKey(addr base.Address) string {
	return fmt.Sprintf("%s:%s", ContractStateKey(addr), DesignStateKeySuffix)
}

var (
	RuntimeStateValueHint = hint.MustNewHint("mitum-contract-runtime-state-value-v0.0.1")
	RuntimeStateKeySuffix = "runtime"

	SnapshotStateValueHint = hint.MustNewHint("mitum-contract-snapshot-state-value-v0.0.1")
	SnapshotStateKeySuffix = "snapshot"
)

type RuntimeEngine string

const (
	RuntimeEngineGnoSnapshot RuntimeEngine = "gno-snapshot-v1"
)

type RuntimeStateValue struct {
	hint.BaseHinter
	Engine          RuntimeEngine
	ABIVersion      string
	PackageName     string
	PackagePath     string
	SnapshotVersion uint64
}

func NewRuntimeStateValue(
	engine RuntimeEngine,
	abiVersion string,
	packageName string,
	packagePath string,
	snapshotVersion uint64,
) RuntimeStateValue {
	return RuntimeStateValue{
		BaseHinter:      hint.NewBaseHinter(RuntimeStateValueHint),
		Engine:          engine,
		ABIVersion:      abiVersion,
		PackageName:     packageName,
		PackagePath:     packagePath,
		SnapshotVersion: snapshotVersion,
	}
}

func (sv RuntimeStateValue) Hint() hint.Hint { return sv.BaseHinter.Hint() }

func (sv RuntimeStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid RuntimeStateValue")

	if err := sv.BaseHinter.IsValid(RuntimeStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}
	if len(sv.Engine) < 1 {
		return e.Errorf("empty engine")
	}
	if len(sv.ABIVersion) < 1 {
		return e.Errorf("empty abi version")
	}
	if len(sv.PackageName) < 1 {
		return e.Errorf("empty package name")
	}
	if len(sv.PackagePath) < 1 {
		return e.Errorf("empty package path")
	}

	return nil
}

func (sv RuntimeStateValue) HashBytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(sv.Engine),
		[]byte(sv.ABIVersion),
		[]byte(sv.PackageName),
		[]byte(sv.PackagePath),
		[]byte(strconv.FormatUint(sv.SnapshotVersion, 10)),
	)
}

func RuntimeStateKey(addr base.Address) string {
	return fmt.Sprintf("%s:%s", ContractStateKey(addr), RuntimeStateKeySuffix)
}

func IsRuntimeStateKey(key string) bool {
	return strings.HasPrefix(key, ContractStateKeyPrefix) && strings.HasSuffix(key, RuntimeStateKeySuffix)
}

func GetRuntimeFromState(st base.State) (RuntimeStateValue, error) {
	v := st.Value()
	if v == nil {
		return RuntimeStateValue{}, common.ErrValueInvalid.Wrap(
			errors.Errorf("state value is nil"),
		)
	}

	sv, ok := v.(RuntimeStateValue)
	if !ok {
		return RuntimeStateValue{}, common.ErrTypeMismatch.Wrap(
			errors.Errorf("expected RuntimeStateValue found, %T", v),
		)
	}

	return sv, nil
}

type SnapshotStateValue struct {
	hint.BaseHinter
	Version  uint64
	Codec    string
	Snapshot []byte
}

func NewSnapshotStateValue(version uint64, codec string, snapshot []byte) SnapshotStateValue {
	return SnapshotStateValue{
		BaseHinter: hint.NewBaseHinter(SnapshotStateValueHint),
		Version:    version,
		Codec:      codec,
		Snapshot:   snapshot,
	}
}

func (sv SnapshotStateValue) Hint() hint.Hint { return sv.BaseHinter.Hint() }

func (sv SnapshotStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid SnapshotStateValue")

	if err := sv.BaseHinter.IsValid(SnapshotStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}
	if len(sv.Codec) < 1 {
		return e.Errorf("empty codec")
	}

	return nil
}

func (sv SnapshotStateValue) HashBytes() []byte {
	return util.ConcatBytesSlice(
		[]byte(strconv.FormatUint(sv.Version, 10)),
		[]byte(sv.Codec),
		valuehash.NewSHA256(sv.Snapshot).Bytes(),
	)
}

func SnapshotStateKey(addr base.Address) string {
	return fmt.Sprintf("%s:%s", ContractStateKey(addr), SnapshotStateKeySuffix)
}

func IsSnapshotStateKey(key string) bool {
	return strings.HasPrefix(key, ContractStateKeyPrefix) && strings.HasSuffix(key, SnapshotStateKeySuffix)
}

func GetSnapshotFromState(st base.State) (SnapshotStateValue, error) {
	v := st.Value()
	if v == nil {
		return SnapshotStateValue{}, common.ErrValueInvalid.Wrap(
			errors.Errorf("state value is nil"),
		)
	}

	sv, ok := v.(SnapshotStateValue)
	if !ok {
		return SnapshotStateValue{}, common.ErrTypeMismatch.Wrap(
			errors.Errorf("expected SnapshotStateValue but %T", v),
		)
	}

	return sv, nil
}
