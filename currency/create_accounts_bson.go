package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
)

func (fact CreateAccountsFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"sender": fact.sender,
				"items":  fact.items,
			},
			fact.BaseFact.BSONM(),
		))
}

type CreateAccountsFactBSONUnmarshaler struct {
	SD string   `bson:"sender"`
	IT bson.Raw `bson:"items"`
}

func (fact *CreateAccountsFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CreateAccountsFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return err
	}

	fact.BaseFact = ubf

	var ucaf CreateAccountsFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &ucaf); err != nil {
		return e(err, "")
	}

	switch a, err := base.DecodeAddress(ucaf.SD, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(ucaf.IT)
	if err != nil {
		return e(err, "")
	}

	items := make([]CreateAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateAccountsItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected CreateAccountsItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}

func (op *CreateAccounts) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
