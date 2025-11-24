package digest

import (
	mongodb "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	bsonutil "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	state "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	types "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type ContractDesignDoc struct {
	mongodb.BaseDoc
	st     base.State
	design types.Design
}

func NewContractDesignDoc(st base.State, enc encoder.Encoder) (ContractDesignDoc, error) {
	design, err := state.GetDesignFromState(st)

	if err != nil {
		return ContractDesignDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, st, enc)
	if err != nil {
		return ContractDesignDoc{}, err
	}

	return ContractDesignDoc{
		BaseDoc: b,
		st:      st,
		design:  design,
	}, nil
}

func (doc ContractDesignDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), state.ContractStateKeyPrefix, 3)

	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()

	return bsonutil.Marshal(m)
}

type ContractDataDoc struct {
	mongodb.BaseDoc
	st   base.State
	data map[string]interface{}
}

func NewContractDataDoc(st base.State, enc encoder.Encoder) (ContractDataDoc, error) {
	data, err := state.GetDataFromState(st)
	if err != nil {
		return ContractDataDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, st, enc)
	if err != nil {
		return ContractDataDoc{}, err
	}

	return ContractDataDoc{
		BaseDoc: b,
		st:      st,
		data:    data,
	}, nil
}

func (doc ContractDataDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), state.ContractStateKeyPrefix, 4)
	if err != nil {
		return nil, err
	}

	m["contract"] = parsedKey[1]
	m["data_key"] = parsedKey[2]
	m["height"] = doc.st.Height()

	return bsonutil.Marshal(m)
}
