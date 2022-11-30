package currency // nolint:dupl

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateAccountsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"keys":    it.keys,
				"amounts": it.amounts,
			}),
	)
}

type CreateAccountsItemBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	KS bson.Raw  `bson:"keys"`
	AM bson.Raw  `bson:"amounts"`
}

func (it *BaseCreateAccountsItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseCreateAccountsItem")

	var uca CreateAccountsItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uca); err != nil {
		return e(err, "")
	}

	it.BaseHinter = hint.NewBaseHinter(uca.HT)

	if hinter, err := enc.Decode(uca.KS); err != nil {
		return err
	} else if k, ok := hinter.(AccountKeys); !ok {
		return e(errors.Errorf("expected AccountsKeys not %T,", hinter), "")
	} else {
		it.keys = k
	}

	ham, err := enc.DecodeSlice(uca.AM)
	if err != nil {
		return err
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return e(util.ErrWrongType.Errorf("expected Amount, not %T", ham[i]), "")
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
