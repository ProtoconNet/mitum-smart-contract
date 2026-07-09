package cmds

import (
	"context"

	"github.com/imfact-labs/currency-model/app/runtime/contracts"
	currencyprocessor "github.com/imfact-labs/currency-model/operation/processor"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	"github.com/imfact-labs/smart-contract-model/operation/processor"
)

var PNameOperationProcessorsMap = ps.Name("mitum-smart-contract-operation-processors-map")

type NewOperationProcessorInternalWithProposalFunc func(base.Height, base.ProposalSignFact, base.GetStateFunc) (base.OperationProcessor, error)

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *currencyprocessor.OperationProcessor
	var setA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var setB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		contracts.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &setA,
		contracts.OperationProcessorsMapBContextKey, &setB,
	); err != nil {
		return pctx, err
	}

	err := opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	if err != nil {
		return pctx, err
	}
	err = opr.SetGetNewProcessorFunc(processor.GetNewProcessor)
	if err != nil {
		return pctx, err
	}
	if err := opr.SetProcessorWithProposal(
		contract.RegisterContractHint,
		contract.NewRegisterContractProcessor(*encs),
	); err != nil {
		return pctx, err
	} else if err := opr.SetProcessorWithProposal(
		contract.CallContractHint,
		contract.NewCallContractProcessor(*encs),
	); err != nil {
		return pctx, err
	}

	_ = setB.Add(contract.RegisterContractHint,
		func(height base.Height, proposal base.ProposalSignFact, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			if err := opr.SetProposal(&proposal); err != nil {
				return nil, err
			}
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	_ = setB.Add(contract.CallContractHint,
		func(height base.Height, proposal base.ProposalSignFact, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			if err := opr.SetProposal(&proposal); err != nil {
				return nil, err
			}
			return opr.New(
				height,
				getStatef,
				nil,
				nil,
			)
		})

	//var f ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc

	pctx = context.WithValue(pctx, contracts.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA)     //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, contracts.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter
	//pctx = context.WithValue(pctx, ProposalOperationFactHintContextKey, f)

	return pctx, nil
}
