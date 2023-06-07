package cmds

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/digest"
	digestisaac "github.com/ProtoconNet/mitum-currency/v3/digest/isaac"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	isaacoperation2 "github.com/ProtoconNet/mitum-currency/v3/operation/isaac"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var hinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: base.BaseStateHint, Instance: base.BaseState{}},
	{Hint: base.NodeHint, Instance: base.BaseNode{}},
	{Hint: base.MEPrivatekeyHint, Instance: base.MEPrivatekey{}},
	{Hint: base.MEPublickeyHint, Instance: base.MEPublickey{}},
	{Hint: base.AccountHint, Instance: base.Account{}},
	{Hint: base.EthAddressHint, Instance: base.EthAddress{}},
	{Hint: base.AddressHint, Instance: base.Address{}},
	{Hint: base.AmountHint, Instance: base.Amount{}},
	{Hint: base.CurrencyDesignHint, Instance: base.CurrencyDesign{}},
	{Hint: base.CurrencyPolicyHint, Instance: base.CurrencyPolicy{}},
	{Hint: base.AccountKeysHint, Instance: base.BaseAccountKeys{}},
	{Hint: base.AccountKeyHint, Instance: base.BaseAccountKey{}},
	{Hint: base.NilFeeerHint, Instance: base.NilFeeer{}},
	{Hint: base.RatioFeeerHint, Instance: base.RatioFeeer{}},
	{Hint: base.FixedFeeerHint, Instance: base.FixedFeeer{}},

	{Hint: base.ContractAccountKeysHint, Instance: base.ContractAccountKeys{}},
	{Hint: extension.CreateContractAccountsItemMultiAmountsHint, Instance: extension.CreateContractAccountsItemMultiAmounts{}},
	{Hint: extension.CreateContractAccountsItemSingleAmountHint, Instance: extension.CreateContractAccountsItemSingleAmount{}},
	{Hint: extension.CreateContractAccountsHint, Instance: extension.CreateContractAccounts{}},
	{Hint: extension.WithdrawsItemMultiAmountsHint, Instance: extension.WithdrawsItemMultiAmounts{}},
	{Hint: extension.WithdrawsItemSingleAmountHint, Instance: extension.WithdrawsItemSingleAmount{}},
	{Hint: extension.WithdrawsHint, Instance: extension.Withdraws{}},

	{Hint: currency.CreateAccountsItemMultiAmountsHint, Instance: currency.CreateAccountsItemMultiAmounts{}},
	{Hint: currency.CreateAccountsItemSingleAmountHint, Instance: currency.CreateAccountsItemSingleAmount{}},
	{Hint: currency.CreateAccountsHint, Instance: currency.CreateAccounts{}},
	{Hint: currency.KeyUpdaterHint, Instance: currency.KeyUpdater{}},
	{Hint: currency.TransfersItemMultiAmountsHint, Instance: currency.TransfersItemMultiAmounts{}},
	{Hint: currency.TransfersItemSingleAmountHint, Instance: currency.TransfersItemSingleAmount{}},
	{Hint: currency.TransfersHint, Instance: currency.Transfers{}},
	{Hint: currency.CurrencyRegisterHint, Instance: currency.CurrencyRegister{}},
	{Hint: currency.CurrencyPolicyUpdaterHint, Instance: currency.CurrencyPolicyUpdater{}},
	{Hint: currency.SuffrageInflationHint, Instance: currency.SuffrageInflation{}},
	{Hint: currency.FeeOperationFactHint, Instance: currency.FeeOperationFact{}},
	{Hint: currency.FeeOperationHint, Instance: currency.FeeOperation{}},
	{Hint: currency.GenesisCurrenciesFactHint, Instance: currency.GenesisCurrenciesFact{}},
	{Hint: currency.GenesisCurrenciesHint, Instance: currency.GenesisCurrencies{}},
	{Hint: statecurrency.AccountStateValueHint, Instance: statecurrency.AccountStateValue{}},
	{Hint: statecurrency.BalanceStateValueHint, Instance: statecurrency.BalanceStateValue{}},
	{Hint: statecurrency.CurrencyDesignStateValueHint, Instance: statecurrency.CurrencyDesignStateValue{}},
	{Hint: digestisaac.ManifestHint, Instance: digestisaac.Manifest{}},
	{Hint: digest.AccountValueHint, Instance: digest.AccountValue{}},
	{Hint: digest.OperationValueHint, Instance: digest.OperationValue{}},
	{Hint: isaacoperation2.GenesisNetworkPolicyHint, Instance: isaacoperation2.GenesisNetworkPolicy{}},
	{Hint: isaacoperation2.SuffrageCandidateHint, Instance: isaacoperation2.SuffrageCandidate{}},
	{Hint: isaacoperation2.SuffrageGenesisJoinHint, Instance: isaacoperation2.SuffrageGenesisJoin{}},
	{Hint: isaacoperation2.SuffrageDisjoinHint, Instance: isaacoperation2.SuffrageDisjoin{}},
	{Hint: isaacoperation2.SuffrageJoinHint, Instance: isaacoperation2.SuffrageJoin{}},
	{Hint: isaacoperation2.NetworkPolicyHint, Instance: isaacoperation2.NetworkPolicy{}},
	{Hint: isaacoperation2.NetworkPolicyStateValueHint, Instance: isaacoperation2.NetworkPolicyStateValue{}},
	{Hint: isaacoperation2.FixedSuffrageCandidateLimiterRuleHint, Instance: isaacoperation2.FixedSuffrageCandidateLimiterRule{}},
	{Hint: isaacoperation2.MajoritySuffrageCandidateLimiterRuleHint, Instance: isaacoperation2.MajoritySuffrageCandidateLimiterRule{}},
}

var supportedProposalOperationFactHinters = []encoder.DecodeDetail{
	{Hint: isaacoperation2.GenesisNetworkPolicyFactHint, Instance: isaacoperation2.GenesisNetworkPolicyFact{}},
	{Hint: isaacoperation2.SuffrageCandidateFactHint, Instance: isaacoperation2.SuffrageCandidateFact{}},
	{Hint: isaacoperation2.SuffrageDisjoinFactHint, Instance: isaacoperation2.SuffrageDisjoinFact{}},
	{Hint: isaacoperation2.SuffrageJoinFactHint, Instance: isaacoperation2.SuffrageJoinFact{}},
	{Hint: isaacoperation2.SuffrageGenesisJoinFactHint, Instance: isaacoperation2.SuffrageGenesisJoinFact{}},
	{Hint: currency.CreateAccountsFactHint, Instance: currency.CreateAccountsFact{}},
	{Hint: currency.KeyUpdaterFactHint, Instance: currency.KeyUpdaterFact{}},
	{Hint: currency.TransfersFactHint, Instance: currency.TransfersFact{}},
	{Hint: currency.CurrencyRegisterFactHint, Instance: currency.CurrencyRegisterFact{}},
	{Hint: currency.CurrencyPolicyUpdaterFactHint, Instance: currency.CurrencyPolicyUpdaterFact{}},
	{Hint: currency.SuffrageInflationFactHint, Instance: currency.SuffrageInflationFact{}},
	{Hint: extension.CreateContractAccountsFactHint, Instance: extension.CreateContractAccountsFact{}},
	{Hint: extension.WithdrawsFactHint, Instance: extension.WithdrawsFact{}},
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
