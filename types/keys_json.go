package types

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/common"

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
	e := util.StringError("failed to decode json of BaseAccountKey")

	var uk KeyJSONUnmarshaler
	if err := enc.Unmarshal(b, &uk); err != nil {
		return e.Wrap(err)
	}

	return ky.unpack(enc, uk.Hint, uk.Weight, uk.Key)
}

type KeysJSONMarshaler struct {
	hint.BaseHinter
	Hash        util.Hash    `json:"hash"`
	Keys        []AccountKey `json:"keys"`
	Threshold   uint         `json:"threshold"`
	AddressType hint.Type    `json:"address_type"`
}

func (ks BaseAccountKeys) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeysJSONMarshaler{
		BaseHinter:  ks.BaseHinter,
		Hash:        ks.h,
		Keys:        ks.keys,
		Threshold:   ks.threshold,
		AddressType: ks.addressType,
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
	Hint        hint.Hint       `json:"_hint"`
	Keys        json.RawMessage `json:"keys"`
	Threshold   uint            `json:"threshold"`
	AddressType string          `json:"address_type"`
}

type KeysHashJSONUnMarshaler struct {
	Hash valuehash.HashDecoder `json:"hash"`
}

type MEKeysHashJSONUnMarshaler struct {
	Hash common.HashDecoder `json:"hash"`
}

func (ks *BaseAccountKeys) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of BaseAccountKeys")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e.Wrap(err)
	}

	var hash util.Hash
	if uks.AddressType == EthAddressHint.Type().String() {
		var uhs MEKeysHashJSONUnMarshaler
		if err := enc.Unmarshal(b, &uhs); err != nil {
			return e.Wrap(err)
		}
		hash = uhs.Hash.Hash()
	} else {
		var uhs KeysHashJSONUnMarshaler
		if err := enc.Unmarshal(b, &uhs); err != nil {
			return e.Wrap(err)
		}
		hash = uhs.Hash.Hash()
	}
	return ks.unpack(enc, uks.Hint, hash, uks.Keys, uks.Threshold, uks.AddressType)
}

func (ks *ContractAccountKeys) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of BaseAccountKeys")

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
