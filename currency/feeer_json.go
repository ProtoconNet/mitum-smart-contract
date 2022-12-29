package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

func (fa NilFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(hint.BaseHinter{
		HT: fa.Hint(),
	})
}

func (fa *NilFeeer) UnmarsahlJSON(b []byte) error {
	var ht hint.BaseHinter
	if err := util.UnmarshalJSON(b, &ht); err != nil {
		return err
	}

	fa.BaseHinter = ht

	return nil
}

type FixedFeeerJSONMarshaler struct {
	hint.BaseHinter
	RC base.Address `json:"receiver"`
	AM string       `json:"amount"`
}

func (fa FixedFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedFeeerJSONMarshaler{
		BaseHinter: fa.BaseHinter,
		RC:         fa.receiver,
		AM:         fa.amount.String(),
	})
}

type FixedFeeerJSONUnmarshaler struct {
	RC string `json:"receiver"`
	AM string `json:"amount"`
}

func (fa *FixedFeeer) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of FixedFeeer")

	var ufa FixedFeeerJSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e(err, "")
	}

	switch ad, err := base.DecodeAddress(ufa.RC, enc); {
	case err != nil:
		return e(err, "")
	default:
		fa.receiver = ad
	}

	if big, err := NewBigFromString(ufa.AM); err != nil {
		return e(err, "")
	} else {
		fa.amount = big
	}

	return nil
}

type RatioFeeerJSONMarshaler struct {
	hint.BaseHinter
	RC base.Address `json:"receiver"`
	RA float64      `json:"ratio"`
	MI string       `json:"min"`
	MA string       `json:"max"`
}

func (fa RatioFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RatioFeeerJSONMarshaler{
		BaseHinter: fa.BaseHinter,
		RC:         fa.receiver,
		RA:         fa.ratio,
		MI:         fa.min.String(),
		MA:         fa.max.String(),
	})
}

type RatioFeeerJSONUnmarshaler struct {
	RC string  `json:"receiver"`
	RA float64 `json:"ratio"`
	MI string  `json:"min"`
	MA string  `json:"max"`
}

func (fa *RatioFeeer) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ufa RatioFeeerJSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return err
	}

	return fa.unpack(enc, ufa.RC, ufa.RA, ufa.MI, ufa.MA)
}
