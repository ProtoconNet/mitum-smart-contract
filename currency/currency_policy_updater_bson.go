package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (fact CurrencyPolicyUpdaterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fact.Hint().String(),
			"currency": fact.currency,
			"policy":   fact.policy,
			"hash":     fact.BaseFact.Hash().String(),
			"token":    fact.BaseFact.Token(),
		},
	)
}

type CurrencyPolicyUpdaterFactBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	CR string   `bson:"currency"`
	PO bson.Raw `bson:"policy"`
}

func (fact *CurrencyPolicyUpdaterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CurrencyPolicyUpdaterFact")

	var u BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf CurrencyPolicyUpdaterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uf.HT)
	if err != nil {
		return e(err, "")
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.CR, uf.PO)
}

func (op CurrencyPolicyUpdater) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
			"memo":  op.Memo,
		})
}

func (op *CurrencyPolicyUpdater) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CurrencyPolicyUpdater")

	var ubo BaseNodeOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	var um MemoBSONUnmarshaler
	if err := enc.Unmarshal(b, &um); err != nil {
		return e(err, "")
	}

	op.BaseNodeOperation = ubo
	op.Memo = um.Memo

	return nil
}
