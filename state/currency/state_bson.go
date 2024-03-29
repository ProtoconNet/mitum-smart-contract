package currency

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (a AccountStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   a.Hint().String(),
			"account": a.Account,
		},
	)
}

type AccountStateValueBSONUnmarshaler struct {
	Hint    string   `bson:"_hint"`
	Account bson.Raw `bson:"account"`
}

func (a *AccountStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode AccountStateValue")

	var u AccountStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	a.BaseHinter = hint.NewBaseHinter(ht)

	var ac types.Account
	if err := ac.DecodeBSON(u.Account, enc); err != nil {
		return e.Wrap(err)
	}

	a.Account = ac

	return nil
}

func (b BalanceStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  b.Hint().String(),
			"amount": b.Amount,
		},
	)
}

type BalanceStateValueBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Amount bson.Raw `bson:"amount"`
}

func (b *BalanceStateValue) DecodeBSON(v []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode BalanceStateValue")

	var u BalanceStateValueBSONUnmarshaler
	if err := enc.Unmarshal(v, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	b.BaseHinter = hint.NewBaseHinter(ht)

	var am types.Amount
	if err := am.DecodeBSON(u.Amount, enc); err != nil {
		return e.Wrap(err)
	}

	b.Amount = am

	return nil
}

func (c CurrencyDesignStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":          c.Hint().String(),
			"currencydesign": c.CurrencyDesign,
		},
	)
}

type CurrencyDesignStateValueBSONUnmarshaler struct {
	Hint           string   `bson:"_hint"`
	CurrencyDesign bson.Raw `bson:"currencydesign"`
}

func (c *CurrencyDesignStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode CurrencyDesignStateValue")

	var u CurrencyDesignStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	c.BaseHinter = hint.NewBaseHinter(ht)

	var cd types.CurrencyDesign
	if err := cd.DecodeBSON(u.CurrencyDesign, enc); err != nil {
		return e.Wrap(err)
	}

	c.CurrencyDesign = cd

	return nil
}
