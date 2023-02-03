package currency

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (s AccountStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   s.Hint().String(),
			"account": s.Account,
		},
	)
}

type AccountStateValueBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	AC bson.Raw `bson:"account"`
}

func (s *AccountStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode AccountStateValue")

	var u AccountStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.HT)
	if err != nil {
		return e(err, "")
	}

	s.BaseHinter = hint.NewBaseHinter(ht)

	var ac Account
	if err := ac.DecodeBSON(u.AC, enc); err != nil {
		return e(err, "")
	}

	s.Account = ac

	return nil
}

func (s BalanceStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  s.Hint().String(),
			"amount": s.Amount,
		},
	)
}

type BalanceStateValueBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	AM bson.Raw `bson:"amount"`
}

func (s *BalanceStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode BalanceStateValue")

	var u BalanceStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.HT)
	if err != nil {
		return e(err, "")
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var am Amount
	if err := am.DecodeBSON(u.AM, enc); err != nil {
		return e(err, "")
	}

	s.Amount = am

	return nil
}

func (s CurrencyDesignStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":          s.Hint().String(),
			"currencydesign": s.CurrencyDesign,
		},
	)
}

type CurrencyDesignStateValueBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	CD bson.Raw `bson:"currencydesign"`
}

func (s *CurrencyDesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyDesignStateValue")

	var u CurrencyDesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.HT)
	if err != nil {
		return e(err, "")
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var cd CurrencyDesign
	if err := cd.DecodeBSON(u.CD, enc); err != nil {
		return e(err, "")
	}

	s.CurrencyDesign = cd

	return nil
}
