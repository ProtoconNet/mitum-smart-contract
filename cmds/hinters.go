package cmds

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum-currency/digest"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util/encoder"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var hinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: currency.AccountHint, Instance: currency.Account{}},
	{Hint: currency.AddressHint, Instance: currency.Address{}},
	{Hint: currency.AmountHint, Instance: currency.Amount{}},
	{Hint: currency.CreateAccountsItemMultiAmountsHint, Instance: currency.CreateAccountsItemMultiAmounts{}},
	{Hint: currency.CreateAccountsItemSingleAmountHint, Instance: currency.CreateAccountsItemSingleAmount{}},
	{Hint: currency.CreateAccountsHint, Instance: currency.CreateAccounts{}},
	{Hint: currency.KeyUpdaterHint, Instance: currency.KeyUpdater{}},
	{Hint: currency.CurrencyDesignHint, Instance: currency.CurrencyDesign{}},
	{Hint: currency.CurrencyPolicyHint, Instance: currency.CurrencyPolicy{}},
	{Hint: currency.CurrencyRegisterHint, Instance: currency.CurrencyRegister{}},
	{Hint: currency.CurrencyPolicyUpdaterHint, Instance: currency.CurrencyPolicyUpdater{}},
	{Hint: currency.TransfersItemMultiAmountsHint, Instance: currency.TransfersItemMultiAmounts{}},
	{Hint: currency.TransfersItemSingleAmountHint, Instance: currency.TransfersItemSingleAmount{}},
	{Hint: currency.TransfersHint, Instance: currency.Transfers{}},
	{Hint: currency.FeeOperationFactHint, Instance: currency.FeeOperationFact{}},
	{Hint: currency.FeeOperationHint, Instance: currency.FeeOperation{}},
	{Hint: currency.FixedFeeerHint, Instance: currency.FixedFeeer{}},
	{Hint: currency.GenesisCurrenciesFactHint, Instance: currency.GenesisCurrenciesFact{}},
	{Hint: currency.GenesisCurrenciesHint, Instance: currency.GenesisCurrencies{}},
	{Hint: currency.AccountKeysHint, Instance: currency.BaseAccountKeys{}},
	{Hint: currency.AccountKeyHint, Instance: currency.BaseAccountKey{}},
	{Hint: currency.NilFeeerHint, Instance: currency.NilFeeer{}},
	{Hint: currency.RatioFeeerHint, Instance: currency.RatioFeeer{}},
	{Hint: currency.AccountStateValueHint, Instance: currency.AccountStateValue{}},
	{Hint: currency.BalanceStateValueHint, Instance: currency.BalanceStateValue{}},
	{Hint: currency.CurrencyDesignStateValueHint, Instance: currency.CurrencyDesignStateValue{}},
	{Hint: digest.AccountValueHint, Instance: digest.AccountValue{}},
	{Hint: digest.OperationValueHint, Instance: digest.OperationValue{}},
}

var supportedProposalOperationFactHinters = []encoder.DecodeDetail{
	{Hint: currency.CreateAccountsFactHint, Instance: currency.CreateAccountsFact{}},
	{Hint: currency.KeyUpdaterFactHint, Instance: currency.KeyUpdaterFact{}},
	{Hint: currency.CurrencyRegisterFactHint, Instance: currency.CurrencyRegisterFact{}},
	{Hint: currency.CurrencyPolicyUpdaterFactHint, Instance: currency.CurrencyPolicyUpdaterFact{}},
	{Hint: currency.TransfersFactHint, Instance: currency.TransfersFact{}},
}

func init() {
	Hinters = make([]encoder.DecodeDetail, len(launch.Hinters)+len(hinters))
	copy(Hinters, launch.Hinters)
	copy(Hinters[len(launch.Hinters):], hinters)

	SupportedProposalOperationFactHinters = make([]encoder.DecodeDetail, len(launch.SupportedProposalOperationFactHinters)+len(supportedProposalOperationFactHinters))
	copy(SupportedProposalOperationFactHinters, launch.SupportedProposalOperationFactHinters)
	copy(SupportedProposalOperationFactHinters[len(launch.SupportedProposalOperationFactHinters):], supportedProposalOperationFactHinters)
}

func LoadHinters(enc encoder.Encoder) error {
	for i := range Hinters {
		if err := enc.Add(Hinters[i]); err != nil {
			return errors.Wrap(err, "failed to add to encoder")
		}
	}

	for i := range SupportedProposalOperationFactHinters {
		if err := enc.Add(SupportedProposalOperationFactHinters[i]); err != nil {
			return errors.Wrap(err, "failed to add to encoder")
		}
	}

	return nil
}
