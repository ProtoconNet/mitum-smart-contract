package currency

import (
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (op BaseOperation) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			op.BaseOperation.BSONM(),
			bson.M{"memo": op.Memo},
		))
}

func (op *BaseOperation) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo base.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	var um MemoBSONUnmarshaler
	if err := enc.Unmarshal(b, &um); err != nil {
		return err
	}
	op.Memo = um.Memo

	return nil
}
