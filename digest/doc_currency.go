package digest

import (
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	mongodbstorage "github.com/ProtoconNet/mitum-currency/v2/digest/mongodb"
	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum-currency/v2/state/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

type CurrencyDoc struct {
	mongodbstorage.BaseDoc
	st base.State
	cd base3.CurrencyDesign
}

// NewBalanceDoc gets the State of Amount
func NewCurrencyDoc(st base.State, enc encoder.Encoder) (CurrencyDoc, error) {
	cd, err := currency.StateCurrencyDesignValue(st)
	if err != nil {
		return CurrencyDoc{}, errors.Wrap(err, "CurrencyDoc needs CurrencyDesign state")
	}

	b, err := mongodbstorage.NewBaseDoc(nil, st, enc)
	if err != nil {
		return CurrencyDoc{}, err
	}

	return CurrencyDoc{
		BaseDoc: b,
		st:      st,
		cd:      cd,
	}, nil
}

func (doc CurrencyDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["currency"] = doc.cd.Currency().String()
	m["height"] = doc.st.Height()

	return bsonenc.Marshal(m)
}
