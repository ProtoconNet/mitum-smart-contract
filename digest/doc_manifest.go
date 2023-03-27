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
	va     base.Manifest
	height base.Height
}

func NewManifestDoc(
	manifest base.Manifest,
	enc encoder.Encoder,
	height base.Height,
	confirmedAt time.Time,
) (ManifestDoc, error) {
	b, err := mongodbstorage.NewBaseDoc(nil, manifest, enc)
	if err != nil {
		return ManifestDoc{}, err
	}

	return ManifestDoc{
		BaseDoc: b,
		va:      manifest,
		height:  height,
	}, nil
}

func (doc ManifestDoc) MarshalBSON() ([]byte, error) {
	m, err := doc.BaseDoc.M()
	if err != nil {
		return nil, err
	}

	m["block"] = doc.va.Hash()
	m["height"] = doc.height

	return bsonenc.Marshal(m)
}
