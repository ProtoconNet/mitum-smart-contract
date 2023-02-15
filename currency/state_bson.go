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
	Hint    string   `bson:"_hint"`
	Account bson.Raw `bson:"account"`
}

func (s *AccountStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode AccountStateValue")

	var u AccountStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}

	s.BaseHinter = hint.NewBaseHinter(ht)

	var ac Account
	if err := ac.DecodeBSON(u.Account, enc); err != nil {
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
	Hint   string   `bson:"_hint"`
	Amount bson.Raw `bson:"amount"`
}

func (s *BalanceStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode BalanceStateValue")

	var u BalanceStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var am Amount
	if err := am.DecodeBSON(u.Amount, enc); err != nil {
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
	Hint           string   `bson:"_hint"`
	CurrencyDesign bson.Raw `bson:"currencydesign"`
}

func (s *CurrencyDesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyDesignStateValue")

	var u CurrencyDesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	var cd CurrencyDesign
	if err := cd.DecodeBSON(u.CurrencyDesign, enc); err != nil {
		return e(err, "")
	}

	s.CurrencyDesign = cd

	return nil
}
