package common

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func (a Big) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.TypeString, bsoncore.AppendString(nil, a.String()), nil
}

func (a *Big) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	if t != bson.TypeString {
		return errors.Errorf("Invalid marshaled type for Big, %v", t)
	}

	s, _, ok := bsoncore.ReadString(b)
	if !ok {
		return errors.Errorf("Can not read string")
	}

	ua, err := NewBigFromString(s)
	if err != nil {
		return err
	}
	*a = ua

	return nil
}
