package isaac

import (
	"time"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (m Manifest) MarshalBSON() ([]byte, error) {
	var previous string
	if m.previous != nil {
		previous = m.previous.String()
	} else {
		previous = ""
	}
	var statesTree string
	if m.statesTree != nil {
		statesTree = m.statesTree.String()
	} else {
		statesTree = ""
	}
	var operationsTree string
	if m.operationsTree != nil {
		statesTree = m.operationsTree.String()
	} else {
		operationsTree = ""
	}
	return bsonenc.Marshal(
		bson.M{
			"_hint":           m.Hint().String(),
			"proposed_at":     m.proposedAt,
			"states_tree":     statesTree,
			"hash":            m.h.String(),
			"previous":        previous,
			"proposal":        m.proposal.String(),
			"operations_tree": operationsTree,
			"suffrage":        m.suffrage.String(),
			"height":          m.height,
		},
	)
}

type ManifestBSONUnmarshaler struct {
	Hint           string      `bson:"_hint"`
	ProposedAt     time.Time   `bson:"proposed_at"`
	StatesTree     string      `bson:"states_tree"`
	Hash           string      `bson:"hash"`
	Previous       string      `bson:"previous"`
	Proposal       string      `bson:"proposal"`
	OperationsTree string      `bson:"operations_tree"`
	Suffrage       string      `bson:"suffrage"`
	Height         base.Height `bson:"height"`
}

func (m *Manifest) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of Manifest")

	var u ManifestBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}

	m.BaseHinter = hint.NewBaseHinter(ht)
	m.h = valuehash.NewBytesFromString(u.Hash)
	m.height = u.Height
	m.previous = valuehash.NewBytesFromString(u.Previous)
	m.proposal = valuehash.NewBytesFromString(u.Proposal)
	m.operationsTree = valuehash.NewBytesFromString(u.OperationsTree)
	m.statesTree = valuehash.NewBytesFromString(u.StatesTree)
	m.suffrage = valuehash.NewBytesFromString(u.Suffrage)
	m.proposedAt = u.ProposedAt

	return nil
}
