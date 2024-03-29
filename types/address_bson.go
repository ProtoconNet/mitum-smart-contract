package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func (ca Address) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.TypeString, bsoncore.AppendString(nil, ca.String()), nil
}

func (ca *Address) DecodeBSON(b []byte, _ *bsonenc.Encoder) error {
	*ca = NewAddress(string(b))

	return nil
}
