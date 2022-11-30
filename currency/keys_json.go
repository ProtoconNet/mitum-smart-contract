package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

type KeyJSONMarshaler struct {
	hint.BaseHinter
	W uint           `json:"weight"`
	K base.Publickey `json:"key"`
}

func (ky BaseAccountKey) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeyJSONMarshaler{
		BaseHinter: ky.BaseHinter,
		W:          ky.w,
		K:          ky.k,
	})
}

type KeyJSONUnmarshaler struct {
	HT hint.Hint `json:"_hint"`
	W  uint      `json:"weight"`
	K  string    `json:"key"`
}

func (ky *BaseAccountKey) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of BaseAccountKey")

	var uk KeyJSONUnmarshaler
	if err := enc.Unmarshal(b, &uk); err != nil {
		return e(err, "")
	}

	return ky.unpack(enc, uk.HT, uk.W, uk.K)
}

type KeysJSONMarshaler struct {
	hint.BaseHinter
	H  util.Hash    `json:"hash"`
	KS []AccountKey `json:"keys"`
	TH uint         `json:"threshold"`
}

func (ks BaseAccountKeys) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeysJSONMarshaler{
		BaseHinter: ks.BaseHinter,
		H:          ks.h,
		KS:         ks.keys,
		TH:         ks.threshold,
	})
}

type KeysJSONUnMarshaler struct {
	HT hint.Hint             `json:"_hint"`
	H  valuehash.HashDecoder `json:"hash"`
	KS json.RawMessage       `json:"keys"`
	TH uint                  `json:"threshold"`
}

func (ks *BaseAccountKeys) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of Account")

	var uks KeysJSONUnMarshaler
	if err := enc.Unmarshal(b, &uks); err != nil {
		return e(err, "")
	}

	ks.BaseHinter = hint.NewBaseHinter(uks.HT)
	hks, err := enc.DecodeSlice(uks.KS)
	if err != nil {
		return err
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return util.ErrWrongType.Errorf("expected Key, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	ks.h = uks.H.Hash()
	ks.threshold = uks.TH

	return nil
}
