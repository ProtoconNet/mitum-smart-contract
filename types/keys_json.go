package types

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

type KeyJSONMarshaler struct {
	hint.BaseHinter
	Weight uint           `json:"weight"`
	Key    base.Publickey `json:"key"`
}

func (ky BaseAccountKey) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeyJSONMarshaler{
		BaseHinter: ky.BaseHinter,
		Weight:     ky.w,
		Key:        ky.k,
	})
}

type KeyJSONUnmarshaler struct {
	Hint   hint.Hint `json:"_hint"`
	Weight uint      `json:"weight"`
	Key    string    `json:"key"`
}

func (ky *BaseAccountKey) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseAccountKey")

	var uk KeyJSONUnmarshaler
	if err := enc.Unmarshal(b, &uk); err != nil {
		return e(err, "")
	}

	return ky.unpack(enc, uk.Hint, uk.Weight, uk.Key)
}

type KeysJSONMarshaler struct {
	hint.BaseHinter
	Hash      util.Hash    `json:"hash"`
	Keys      []AccountKey `json:"keys"`
	Threshold uint         `json:"threshold"`
}

func (ks BaseAccountKeys) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeysJSONMarshaler{
		BaseHinter: ks.BaseHinter,
		Hash:       ks.h,
		Keys:       ks.keys,
		Threshold:  ks.threshold,
	})
}

type KeysJSONUnMarshaler struct {
	Hint      hint.Hint             `json:"_hint"`
	Hash      valuehash.HashDecoder `json:"hash"`
	Keys      json.RawMessage       `json:"keys"`
	Threshold uint                  `json:"threshold"`
}

func (ks *BaseAccountKeys) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseAccountKeys")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e(err, "")
	}

	return ks.unpack(enc, uks.Hint, uks.Hash, uks.Keys, uks.Threshold)

}
