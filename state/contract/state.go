package contract

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util/valuehash"

	types "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
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
}

func NewDesignStateValue(design types.Design) DesignStateValue {
	return DesignStateValue{
		BaseHinter: hint.NewBaseHinter(DesignStateValueHint),
		Design:     design,
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

	return nil
}

func (sv DesignStateValue) HashBytes() []byte {
	return sv.Design.Bytes()
}

func GetDesignFromState(st base.State) (types.Design, error) {
	v := st.Value()
	if v == nil {
		return types.Design{}, errors.Errorf("state value is nil")
	}

	d, ok := v.(DesignStateValue)
	if !ok {
		return types.Design{}, errors.Errorf("expected DesignStateValue but %T", v)
	}

	return d.Design, nil
}

func IsDesignStateKey(key string) bool {
	return strings.HasPrefix(key, ContractStateKeyPrefix) && strings.HasSuffix(key, DesignStateKeySuffix)
}

func DesignStateKey(addr base.Address) string {
	return fmt.Sprintf("%s:%s", ContractStateKey(addr), DesignStateKeySuffix)
}

var (
	DataStateValueHint = hint.MustNewHint("mitum-contract-data-state-value-v0.0.1")
	DataStateKeySuffix = "data"
)

type DataStateValue struct {
	hint.BaseHinter
	Data map[string]interface{}
}

func NewDataStateValue(data map[string]interface{}) DataStateValue {
	return DataStateValue{
		BaseHinter: hint.NewBaseHinter(DataStateValueHint),
		Data:       data,
	}
}

func (sv DataStateValue) Hint() hint.Hint {
	return sv.BaseHinter.Hint()
}

func (sv DataStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid DataStateValue")

	if err := sv.BaseHinter.IsValid(DataStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if sv.Data != nil {
		if _, err := json.Marshal(sv.Data); err != nil {
			return e.Wrap(errors.Wrap(err, "data is not JSON-serializable"))
		}
	}

	return nil
}

func (sv DataStateValue) HashBytes() []byte {
	var bs [][]byte
	if sv.Data != nil {
		d, _ := json.Marshal(sv.Data)
		bs = append(bs, valuehash.NewSHA256(d).Bytes())
	}

	return util.ConcatBytesSlice(bs...)
}

func GetDataFromState(st base.State) (map[string]interface{}, error) {
	v := st.Value()
	if v == nil {
		return nil, errors.Errorf("State value is nil")
	}

	ts, ok := v.(DataStateValue)
	if !ok {
		return nil, common.ErrTypeMismatch.Wrap(errors.Errorf("expected DataStateValue found, %T", v))
	}

	return ts.Data, nil
}

func IsDataStateKey(key string) bool {
	return strings.HasPrefix(key, ContractStateKeyPrefix) && strings.HasSuffix(key, DataStateKeySuffix)
}

func DataStateKey(addr base.Address, key string) string {
	return fmt.Sprintf("%s:%s:%s", ContractStateKey(addr), key, DataStateKeySuffix)
}

var (
	RuntimeStateValueHint = hint.MustNewHint("mitum-contract-runtime-state-value-v0.0.1")
	RuntimeStateKeySuffix = "runtime"

	SnapshotStateValueHint = hint.MustNewHint("mitum-contract-snapshot-state-value-v0.0.1")
	SnapshotStateKeySuffix = "snapshot"
)

type RuntimeEngine string

const (
	RuntimeEngineYaegi       RuntimeEngine = "yaegi-v1"
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
