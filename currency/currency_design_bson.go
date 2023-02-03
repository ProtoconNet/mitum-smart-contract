package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

func (de CurrencyDesign) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":           de.Hint().String(),
			"amount":          de.amount,
			"genesis_account": de.genesisAccount,
			"policy":          de.policy,
			"aggregate":       de.aggregate.String(),
		},
	)
}

type CurrencyDesignBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	AM bson.Raw `bson:"amount"`
	GA string   `bson:"genesis_account"`
	PO bson.Raw `bson:"policy"`
	AG string   `bson:"aggregate"`
}

func (de *CurrencyDesign) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CurrencyDesign")

	var ude CurrencyDesignBSONUnmarshaler
	if err := enc.Unmarshal(b, &ude); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(ude.HT)
	if err != nil {
		return e(err, "")
	}

	return de.unpack(enc, ht, ude.AM, ude.GA, ude.PO, ude.AG)
}
