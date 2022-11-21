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
	AM Big          `json:"amount"`
}

func (fa FixedFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FixedFeeerJSONMarshaler{
		BaseHinter: fa.BaseHinter,
		RC:         fa.receiver,
		AM:         fa.amount,
	})
}

type FixedFeeerJSONUnmarshaler struct {
	HT hint.Hint `json:"_hint"`
	RC string    `json:"receiver"`
	AM Big       `json:"amount"`
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
	fa.amount = ufa.AM

	return nil
}

type RatioFeeerJSONMarshaler struct {
	hint.BaseHinter
	RC base.Address `json:"receiver"`
	RA float64      `json:"ratio"`
	MI Big          `json:"min"`
	MA Big          `json:"max"`
}

func (fa RatioFeeer) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RatioFeeerJSONMarshaler{
		BaseHinter: fa.BaseHinter,
		RC:         fa.receiver,
		RA:         fa.ratio,
		MI:         fa.min,
		MA:         fa.max,
	})
}

type RatioFeeerJSONUnmarshaler struct {
	RC string  `json:"receiver"`
	RA float64 `json:"ratio"`
	MI Big     `json:"min"`
	MA Big     `json:"max"`
}

func (fa *RatioFeeer) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of RatioFeeer")

	var ufa RatioFeeerJSONUnmarshaler
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return err
	}

	fa.ratio = ufa.RA
	fa.max = ufa.MA
	fa.min = ufa.MI

	switch ad, err := base.DecodeAddress(ufa.RC, enc); {
	case err != nil:
		return e(err, "")
	default:
		fa.receiver = ad
	}

	return nil
}
