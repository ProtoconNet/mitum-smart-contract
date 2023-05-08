package digest

import (
	"time"

	mongodbstorage "github.com/ProtoconNet/mitum-currency/v2/digest/mongodb"
	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type ManifestDoc struct {
	mongodbstorage.BaseDoc
	va         base.Manifest
	operations []base.Operation
	height     base.Height
}

func NewManifestDoc(
	manifest base.Manifest,
	enc encoder.Encoder,
	height base.Height,
	operations []base.Operation,
	confirmedAt time.Time,
) (ManifestDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, manifest, enc)
	if err != nil {
		return ManifestDoc{}, err
	}

	return ManifestDoc{
		BaseDoc:    b,
		va:         manifest,
		operations: operations,
		height:     height,
	}, nil
}

func (doc ManifestDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["block"] = doc.va.Hash()
	m["operations"] = len(doc.operations)
	m["height"] = doc.height

	return bsonenc.Marshal(m)
}
