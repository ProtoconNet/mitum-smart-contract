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
	var contractDataModels []mongo.WriteModel
	for i := range bs.sts {
		st := bs.sts[i]
		switch {
		case pstate.IsDesignStateKey(st.Key()):
			j, err := bs.handleContractDesignState(st)
			if err != nil {
				return err
			}
			contractModels = append(contractModels, j...)
		case pstate.IsDataStateKey(st.Key()):
			j, err := bs.handleContractDataState(st)
			if err != nil {
				return err
			}
			contractDataModels = append(contractDataModels, j...)
		default:
			continue
		}
	}

	bs.contractModels = contractModels
	bs.contractDataModels = contractDataModels

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

func (bs *BlockSession) handleContractDataState(st base.State) ([]mongo.WriteModel, error) {
	if dataDoc, err := NewContractDataDoc(st, bs.st.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{
			mongo.NewInsertOneModel().SetDocument(dataDoc),
		}, nil
	}
}
