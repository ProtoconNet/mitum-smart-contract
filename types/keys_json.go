package types

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
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

func (ky *BaseAccountKey) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of BaseAccountKey")

	var uk KeyJSONUnmarshaler
	if err := enc.Unmarshal(b, &uk); err != nil {
		return e.Wrap(err)
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

func (ks NilAccountKeys) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeysJSONMarshaler{
		BaseHinter: ks.BaseHinter,
		Hash:       ks.h,
		Threshold:  ks.threshold,
	})
}

func (ks ContractAccountKeys) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeysJSONMarshaler{
		BaseHinter: ks.BaseHinter,
		Hash:       ks.h,
		Keys:       ks.keys,
		Threshold:  ks.threshold,
	})
}

type KeysJSONUnMarshaler struct {
	Hint      hint.Hint       `json:"_hint"`
	Keys      json.RawMessage `json:"keys"`
	Threshold uint            `json:"threshold"`
}

type KeysHashJSONUnMarshaler struct {
	Hash common.HashDecoder `json:"hash"`
}

func (ks *BaseAccountKeys) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of BaseAccountKeys")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e.Wrap(err)
	}

	var hash util.Hash
	var uhs KeysHashJSONUnMarshaler
	if err := enc.Unmarshal(b, &uhs); err != nil {
		return e.Wrap(err)
	}
	hash = uhs.Hash.Hash()

	return ks.unpack(enc, uks.Hint, hash, uks.Keys, uks.Threshold)
}

func (ks *NilAccountKeys) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of NilAccountKeys")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e.Wrap(err)
	}

	var hash util.Hash
	var uhs KeysHashJSONUnMarshaler
	if err := enc.Unmarshal(b, &uhs); err != nil {
		return e.Wrap(err)
	}
	hash = uhs.Hash.Hash()

	return ks.unpack(enc, uks.Hint, hash, uks.Threshold)
}

func (ks *ContractAccountKeys) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("Decode json of BaseAccountKeys")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e.Wrap(err)
	}

	var uhs KeysHashJSONUnMarshaler
	if err := enc.Unmarshal(b, &uhs); err != nil {
		return e.Wrap(err)
	}

	return ks.unpack(enc, uks.Hint, uhs.Hash, uks.Keys, uks.Threshold)

}
