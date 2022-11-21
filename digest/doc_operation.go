package digest

import (
	"time"

	"github.com/spikeekips/mitum-currency/currency"
	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
	mongodbstorage "github.com/spikeekips/mitum-currency/digest/mongodb"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util/encoder"
)

type OperationDoc struct {
	mongodbstorage.BaseDoc
	va        OperationValue
	op        base.Operation
	addresses []string
	height    base.Height
}

func NewOperationDoc(
	op base.Operation,
	enc encoder.Encoder,
	height base.Height,
	confirmedAt time.Time,
	inState bool,
	reason base.OperationProcessReasonError,
	index uint64,
) (OperationDoc, error) {
	var addresses []string
	if ads, ok := op.Fact().(currency.Addresses); ok {
		as, err := ads.Addresses()
		if err != nil {
			return OperationDoc{}, err
		}
		addresses = make([]string, len(as))
		for i := range as {
			addresses[i] = as[i].String()
		}
	}

	va := NewOperationValue(op, height, confirmedAt, inState, reason, index)
	b, err := mongodbstorage.NewBaseDoc(nil, va, enc)
	if err != nil {
		return OperationDoc{}, err
	}

	return OperationDoc{
		BaseDoc:   b,
		va:        va,
		op:        op,
		addresses: addresses,
		height:    height,
	}, nil
}

func (doc OperationDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["addresses"] = doc.addresses
	m["fact"] = doc.op.Fact().Hash()
	m["height"] = doc.height
	m["index"] = doc.va.index

	return bsonenc.Marshal(m)
}
