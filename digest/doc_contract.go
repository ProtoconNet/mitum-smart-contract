package digest

import (
	"crypto/sha256"
	"encoding/hex"

	mongodb "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	bsonutil "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
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
	design, err := pstate.GetDesignFromState(st)

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

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), pstate.ContractStateKeyPrefix, 3)
	if err != nil {
		return nil, err
	}

	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()

	return bsonutil.Marshal(m)
}

type ContractRuntimeDoc struct {
	mongodb.BaseDoc
	st      base.State
	runtime pstate.RuntimeStateValue
}

func NewContractRuntimeDoc(st base.State, enc encoder.Encoder) (ContractRuntimeDoc, error) {
	runtimeValue, err := pstate.GetRuntimeFromState(st)
	if err != nil {
		return ContractRuntimeDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, st, enc)
	if err != nil {
		return ContractRuntimeDoc{}, err
	}

	return ContractRuntimeDoc{
		BaseDoc: b,
		st:      st,
		runtime: runtimeValue,
	}, nil
}

func (doc ContractRuntimeDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), pstate.ContractStateKeyPrefix, 3)
	if err != nil {
		return nil, err
	}

	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()
	m["engine"] = doc.runtime.Engine
	m["abi_version"] = doc.runtime.ABIVersion
	m["package_name"] = doc.runtime.PackageName
	m["package_path"] = doc.runtime.PackagePath
	m["snapshot_version"] = doc.runtime.SnapshotVersion

	return bsonutil.Marshal(m)
}

type ContractSnapshotDoc struct {
	mongodb.BaseDoc
	st       base.State
	snapshot pstate.SnapshotStateValue
}

func NewContractSnapshotDoc(st base.State, enc encoder.Encoder) (ContractSnapshotDoc, error) {
	snapshotValue, err := pstate.GetSnapshotFromState(st)
	if err != nil {
		return ContractSnapshotDoc{}, err
	}

	b, err := mongodb.NewBaseDoc(nil, st, enc)
	if err != nil {
		return ContractSnapshotDoc{}, err
	}

	return ContractSnapshotDoc{
		BaseDoc:  b,
		st:       st,
		snapshot: snapshotValue,
	}, nil
}

func (doc ContractSnapshotDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	parsedKey, err := cstate.ParseStateKey(doc.st.Key(), pstate.ContractStateKeyPrefix, 3)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(doc.snapshot.Snapshot)

	m["contract"] = parsedKey[1]
	m["height"] = doc.st.Height()
	m["version"] = doc.snapshot.Version
	m["codec"] = doc.snapshot.Codec
	m["snapshot_size"] = len(doc.snapshot.Snapshot)
	m["snapshot_sha256"] = hex.EncodeToString(sum[:])

	return bsonutil.Marshal(m)
}
