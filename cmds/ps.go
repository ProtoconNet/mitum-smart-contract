package cmds

import (
	"context"

	ccmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	currencyprocessor "github.com/ProtoconNet/mitum-currency/v3/operation/processor"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract"
	"github.com/ProtoconNet/mitum-smart-contract/operation/processor"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/ps"
)

var PNameOperationProcessorsMap = ps.Name("mitum-smart-contract-operation-processors-map")

type NewOperationProcessorInternalWithProposalFunc func(base.Height, base.ProposalSignFact, base.GetStateFunc) (base.OperationProcessor, error)

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var isaacParams *isaac.Params
	var db isaac.Database
	var opr *currencyprocessor.OperationProcessor
	var setA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var setB *hint.CompatibleSet[ccmds.NewOperationProcessorInternalWithProposalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacParams,
		launch.CenterDatabaseContextKey, &db,
		ccmds.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &setA,
		ccmds.OperationProcessorsMapBContextKey, &setB,
	); err != nil {
		return pctx, err
	}

	//err = opr.SetCheckDuplicationFunc(processor.CheckDuplication)
	//if err != nil {
	//	return pctx, err
	//}
	err := opr.SetGetNewProcessorFunc(processor.GetNewProcessor)
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

	pctx = context.WithValue(pctx, ccmds.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA) //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, ccmds.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter
	//pctx = context.WithValue(pctx, ProposalOperationFactHintContextKey, f)

	return pctx, nil
}
