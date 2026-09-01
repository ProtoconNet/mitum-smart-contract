package steps

import (
	"context"

	csteps "github.com/imfact-labs/currency-model/app/runtime/steps"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/smart-contract-model/runtime/contracts"
	"github.com/imfact-labs/smart-contract-model/runtime/spec"
	"github.com/pkg/errors"
)

func PAddHinters(pctx context.Context) (context.Context, error) {
	e := util.StringError("add hinters")
	var encs *encoder.Encoders
	var f contracts.ProposalOperationFactHintFunc = IsSupportedProposalOperationFactHintFunc
	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &encs); err != nil {
		return pctx, e.Wrap(err)
	}
	pctx = context.WithValue(pctx, contracts.ProposalOperationFactHintContextKey, f)
	if err := LoadHinters(encs); err != nil {
		return pctx, e.Wrap(err)
	}
	return pctx, nil
}

func LoadHinters(encs *encoder.Encoders) error {
	if err := csteps.LoadHinters(encs); err != nil {
		return err
	}
	for i := range spec.AddedHinters {
		if err := encs.AddDetail(spec.AddedHinters[i]); err != nil {
			return errors.Wrap(err, "add smart-contract hinter to encoder")
		}
	}
	for i := range spec.AddedSupportedHinters {
		if err := encs.AddDetail(spec.AddedSupportedHinters[i]); err != nil {
			return errors.Wrap(err, "add smart-contract supported fact hinter to encoder")
		}
	}
	return nil
}

func IsSupportedProposalOperationFactHintFunc() func(hint.Hint) bool {
	currencyFilter := csteps.IsSupportedProposalOperationFactHintFunc()
	return func(ht hint.Hint) bool {
		if currencyFilter(ht) {
			return true
		}
		for i := range spec.AddedSupportedHinters {
			s := spec.AddedSupportedHinters[i].Hint
			if ht.Type() == s.Type() {
				return ht.IsCompatible(s)
			}
		}
		return false
	}
}
