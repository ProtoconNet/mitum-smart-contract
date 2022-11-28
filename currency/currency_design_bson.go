package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
)

func (de CurrencyDesign) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(de.Hint()),
		bson.M{
			"amount":          de.amount,
			"genesis_account": de.genesisAccount,
			"policy":          de.policy,
			"aggregate":       de.aggregate,
		}),
	)
}

type CurrencyDesignBSONUnmarshaler struct {
	AM bson.Raw `bson:"amount"`
	GA string   `bson:"genesis_account"`
	PO bson.Raw `bson:"policy"`
	AG Big      `bson:"aggregate"`
}

func (de *CurrencyDesign) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ude CurrencyDesignBSONUnmarshaler
	if err := enc.Unmarshal(b, &ude); err != nil {
		return err
	}

	return de.unpack(enc, ude.AM, ude.GA, ude.PO, ude.AG)
}
