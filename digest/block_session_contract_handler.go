package digest

import (
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"go.mongodb.org/mongo-driver/mongo"
)

func (bs *BlockSession) prepareContract() error {
	if len(bs.sts) < 1 {
		return nil
	}

	var contractModels []mongo.WriteModel
	var contractRuntimeModels []mongo.WriteModel
	var contractSnapshotModels []mongo.WriteModel

	for i := range bs.sts {
		st := bs.sts[i]
		switch {
		case pstate.IsDesignStateKey(st.Key()):
			j, err := bs.handleContractDesignState(st)
			if err != nil {
				return err
			}
			contractModels = append(contractModels, j...)

		case pstate.IsRuntimeStateKey(st.Key()):
			j, err := bs.handleContractRuntimeState(st)
			if err != nil {
				return err
			}
			contractRuntimeModels = append(contractRuntimeModels, j...)

		case pstate.IsSnapshotStateKey(st.Key()):
			j, err := bs.handleContractSnapshotState(st)
			if err != nil {
				return err
			}
			contractSnapshotModels = append(contractSnapshotModels, j...)

		default:
			continue
		}
	}

	bs.contractModels = contractModels
	bs.contractRuntimeModels = contractRuntimeModels
	bs.contractSnapshotModels = contractSnapshotModels

	return nil
}

func (bs *BlockSession) handleContractDesignState(st base.State) ([]mongo.WriteModel, error) {
	if designDoc, err := NewContractDesignDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(designDoc),
		}, nil
	}
}

func (bs *BlockSession) handleContractRuntimeState(st base.State) ([]mongo.WriteModel, error) {
	if doc, err := NewContractRuntimeDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(doc),
		}, nil
	}
}

func (bs *BlockSession) handleContractSnapshotState(st base.State) ([]mongo.WriteModel, error) {
	if doc, err := NewContractSnapshotDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(doc),
		}, nil
	}
}
