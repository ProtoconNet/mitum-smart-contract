package steps

import (
	"context"

	currencyprocessor "github.com/imfact-labs/currency-model/operation/processor"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	"github.com/imfact-labs/smart-contract-model/operation/processor"
	"github.com/imfact-labs/smart-contract-model/runtime/contracts"
)

var PNameOperationProcessorsMap = ps.Name("mitum-smart-contract-operation-processors-map")

func OperationProcessorHints() []hint.Hint {
	return []hint.Hint{contract.RegisterContractHint, contract.CallContractHint}
}

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	var encs *encoder.Encoders
	var opr *currencyprocessor.OperationProcessor
	var setA *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var setB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc]
	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		contracts.OperationProcessorContextKey, &opr,
		launch.OperationProcessorsMapContextKey, &setA,
		contracts.OperationProcessorsMapBContextKey, &setB,
	); err != nil {
		return pctx, err
	}
	if err := opr.SetCheckDuplicationFunc(processor.CheckDuplication); err != nil {
		return pctx, err
	}
	if err := opr.SetGetNewProcessorFunc(processor.GetNewProcessor); err != nil {
		return pctx, err
	}
	if err := opr.SetProcessorWithProposal(contract.RegisterContractHint, contract.NewRegisterContractProcessor(*encs)); err != nil {
		return pctx, err
	}
	if err := opr.SetProcessorWithProposal(contract.CallContractHint, contract.NewCallContractProcessor(*encs)); err != nil {
		return pctx, err
	}
	for _, operationHint := range OperationProcessorHints() {
		if err := setB.Add(operationHint, func(height base.Height, proposal base.ProposalSignFact, getStatef base.GetStateFunc) (base.OperationProcessor, error) {
			if err := opr.SetProposal(&proposal); err != nil {
				return nil, err
			}
			return opr.New(height, getStatef, nil, nil)
		}); err != nil {
			return pctx, err
		}
	}
	pctx = context.WithValue(pctx, contracts.OperationProcessorContextKey, opr)
	pctx = context.WithValue(pctx, launch.OperationProcessorsMapContextKey, setA)     //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, contracts.OperationProcessorsMapBContextKey, setB) //revive:disable-line:modifies-parameter
	return pctx, nil
}
