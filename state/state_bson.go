package state

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-smart-contract/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (sv DesignStateValue) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint":  sv.Hint().String(),
		"design": sv.Design,
	}
	if sv.Schema != nil {
		m["schema"] = sv.Schema
	}

	return bsonenc.Marshal(m)
}

type DesignStateValueBSONUnmarshaler struct {
	Hint   string                         `bson:"_hint"`
	Design bson.Raw                       `bson:"design"`
	Schema *types.PersistedContractSchema `bson:"schema"`
}

func (sv *DesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of DesignStateValue")

	var u DesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	sv.BaseHinter = hint.NewBaseHinter(ht)

	var sd types.Design
	if err := sd.DecodeBSON(u.Design, enc); err != nil {
		return e.Wrap(err)
	}
	sv.Design = sd
	sv.Schema = u.Schema

	return nil
}

func (sv RuntimeStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":            sv.Hint().String(),
		"engine":           sv.Engine,
		"abi_version":      sv.ABIVersion,
		"package_name":     sv.PackageName,
		"package_path":     sv.PackagePath,
		"snapshot_version": sv.SnapshotVersion,
	})
}

type RuntimeStateValueBSONUnmarshaler struct {
	Hint            string        `bson:"_hint"`
	Engine          RuntimeEngine `bson:"engine"`
	ABIVersion      string        `bson:"abi_version"`
	PackageName     string        `bson:"package_name"`
	PackagePath     string        `bson:"package_path"`
	SnapshotVersion uint64        `bson:"snapshot_version"`
}

func (sv *RuntimeStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u RuntimeStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return err
	}

	sv.BaseHinter = hint.NewBaseHinter(ht)
	sv.Engine = u.Engine
	sv.ABIVersion = u.ABIVersion
	sv.PackageName = u.PackageName
	sv.PackagePath = u.PackagePath
	sv.SnapshotVersion = u.SnapshotVersion

	return nil
}

func (sv SnapshotStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bson.M{
		"_hint":    sv.Hint().String(),
		"version":  sv.Version,
		"codec":    sv.Codec,
		"snapshot": sv.Snapshot,
	})
}

type SnapshotStateValueBSONUnmarshaler struct {
	Hint     string `bson:"_hint"`
	Version  uint64 `bson:"version"`
	Codec    string `bson:"codec"`
	Snapshot []byte `bson:"snapshot"`
}

func (sv *SnapshotStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u SnapshotStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return err
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return err
	}

	sv.BaseHinter = hint.NewBaseHinter(ht)
	sv.Version = u.Version
	sv.Codec = u.Codec
	sv.Snapshot = u.Snapshot

	return nil
}
