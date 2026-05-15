package contract

import (
	"encoding/json"

	types "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type DesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	Design types.Design `json:"design"`
}

func (sv DesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DesignStateValueJSONMarshaler(sv),
	)
}

type DesignStateValueJSONUnmarshaler struct {
	Hint   hint.Hint       `json:"_hint"`
	Design json.RawMessage `json:"design"`
}

func (sv *DesignStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("failed to decode json of DesignStateValue")

	var u DesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)

	var sd types.Design
	if err := sd.DecodeJSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Design = sd

	return nil
}

type DataStateValueJSONMarshaler struct {
	hint.BaseHinter
	Data map[string]interface{} `json:"data"`
}

func (sv DataStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(
		DataStateValueJSONMarshaler(sv),
	)
}

type DataStateValueJSONUnmarshaler struct {
	Hint hint.Hint              `json:"_hint"`
	Data map[string]interface{} `json:"data"`
}

func (sv *DataStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of DataStateValue")

	var u DataStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)
	sv.Data = u.Data

	return nil
}

type RuntimeStateValueJSONMarshaler struct {
	hint.BaseHinter
	Engine          RuntimeEngine `json:"engine"`
	ABIVersion      string        `json:"abi_version"`
	PackageName     string        `json:"package_name"`
	PackagePath     string        `json:"package_path"`
	SnapshotVersion uint64        `json:"snapshot_version"`
}

func (sv RuntimeStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RuntimeStateValueJSONMarshaler(sv))
}

type RuntimeStateValueJSONUnmarshaler struct {
	Hint            hint.Hint     `json:"_hint"`
	Engine          RuntimeEngine `json:"engine"`
	ABIVersion      string        `json:"abi_version"`
	PackageName     string        `json:"package_name"`
	PackagePath     string        `json:"package_path"`
	SnapshotVersion uint64        `json:"snapshot_version"`
}

func (sv *RuntimeStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u RuntimeStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)
	sv.Engine = u.Engine
	sv.ABIVersion = u.ABIVersion
	sv.PackageName = u.PackageName
	sv.PackagePath = u.PackagePath
	sv.SnapshotVersion = u.SnapshotVersion

	return nil
}

type SnapshotStateValueJSONMarshaler struct {
	hint.BaseHinter
	Version  uint64 `json:"version"`
	Codec    string `json:"codec"`
	Snapshot []byte `json:"snapshot"`
}

func (sv SnapshotStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SnapshotStateValueJSONMarshaler(sv))
}

type SnapshotStateValueJSONUnmarshaler struct {
	Hint     hint.Hint `json:"_hint"`
	Version  uint64    `json:"version"`
	Codec    string    `json:"codec"`
	Snapshot []byte    `json:"snapshot"`
}

func (sv *SnapshotStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var u SnapshotStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	sv.BaseHinter = hint.NewBaseHinter(u.Hint)
	sv.Version = u.Version
	sv.Codec = u.Codec
	sv.Snapshot = u.Snapshot

	return nil
}
