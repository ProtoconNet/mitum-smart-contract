package cmds

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/digest"
	digestisaac "github.com/ProtoconNet/mitum-currency/v3/digest/isaac"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	isaacoperation "github.com/ProtoconNet/mitum-currency/v3/operation/isaac"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var AddedHinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: common.BaseStateHint, Instance: common.BaseState{}},
	{Hint: common.NodeHint, Instance: common.BaseNode{}},

	{Hint: types.AccountHint, Instance: types.Account{}},
	{Hint: types.AccountKeyHint, Instance: types.BaseAccountKey{}},
	{Hint: types.AccountKeysHint, Instance: types.BaseAccountKeys{}},
	{Hint: types.AddressHint, Instance: types.Address{}},
	{Hint: types.AmountHint, Instance: types.Amount{}},
	{Hint: types.ContractAccountKeysHint, Instance: types.ContractAccountKeys{}},
	{Hint: types.CurrencyDesignHint, Instance: types.CurrencyDesign{}},
	{Hint: types.CurrencyPolicyHint, Instance: types.CurrencyPolicy{}},
	{Hint: types.EthAddressHint, Instance: types.EthAddress{}},
	{Hint: types.FixedFeeerHint, Instance: types.FixedFeeer{}},
	{Hint: types.MEPrivatekeyHint, Instance: types.MEPrivatekey{}},
	{Hint: types.MEPublickeyHint, Instance: types.MEPublickey{}},
	{Hint: types.NilFeeerHint, Instance: types.NilFeeer{}},
	{Hint: types.RatioFeeerHint, Instance: types.RatioFeeer{}},

	{Hint: currency.CreateAccountsHint, Instance: currency.CreateAccounts{}},
	{Hint: currency.CreateAccountsItemMultiAmountsHint, Instance: currency.CreateAccountsItemMultiAmounts{}},
	{Hint: currency.CreateAccountsItemSingleAmountHint, Instance: currency.CreateAccountsItemSingleAmount{}},
	{Hint: currency.CurrencyPolicyUpdaterHint, Instance: currency.CurrencyPolicyUpdater{}},
	{Hint: currency.CurrencyRegisterHint, Instance: currency.CurrencyRegister{}},
	//{Hint: currency.FeeOperationFactHint, Instance: currency.FeeOperationFact{}},
	//{Hint: currency.FeeOperationHint, Instance: currency.FeeOperation{}},
	{Hint: currency.GenesisCurrenciesHint, Instance: currency.GenesisCurrencies{}},
	{Hint: currency.GenesisCurrenciesFactHint, Instance: currency.GenesisCurrenciesFact{}},
	{Hint: currency.KeyUpdaterHint, Instance: currency.KeyUpdater{}},
	{Hint: currency.SuffrageInflationHint, Instance: currency.SuffrageInflation{}},
	{Hint: currency.TransfersHint, Instance: currency.Transfers{}},
	{Hint: currency.TransfersItemMultiAmountsHint, Instance: currency.TransfersItemMultiAmounts{}},
	{Hint: currency.TransfersItemSingleAmountHint, Instance: currency.TransfersItemSingleAmount{}},

	{Hint: extension.CreateContractAccountsHint, Instance: extension.CreateContractAccounts{}},
	{Hint: extension.CreateContractAccountsItemMultiAmountsHint, Instance: extension.CreateContractAccountsItemMultiAmounts{}},
	{Hint: extension.CreateContractAccountsItemSingleAmountHint, Instance: extension.CreateContractAccountsItemSingleAmount{}},
	{Hint: extension.WithdrawsHint, Instance: extension.Withdraws{}},
	{Hint: extension.WithdrawsItemMultiAmountsHint, Instance: extension.WithdrawsItemMultiAmounts{}},
	{Hint: extension.WithdrawsItemSingleAmountHint, Instance: extension.WithdrawsItemSingleAmount{}},

	{Hint: isaacoperation.GenesisNetworkPolicyHint, Instance: isaacoperation.GenesisNetworkPolicy{}},
	{Hint: isaacoperation.FixedSuffrageCandidateLimiterRuleHint, Instance: isaacoperation.FixedSuffrageCandidateLimiterRule{}},
	{Hint: isaacoperation.MajoritySuffrageCandidateLimiterRuleHint, Instance: isaacoperation.MajoritySuffrageCandidateLimiterRule{}},
	{Hint: isaacoperation.NetworkPolicyHint, Instance: isaacoperation.NetworkPolicy{}},
	{Hint: isaacoperation.NetworkPolicyStateValueHint, Instance: isaacoperation.NetworkPolicyStateValue{}},
	{Hint: isaacoperation.SuffrageCandidateHint, Instance: isaacoperation.SuffrageCandidate{}},
	{Hint: isaacoperation.SuffrageDisjoinHint, Instance: isaacoperation.SuffrageDisjoin{}},
	{Hint: isaacoperation.SuffrageGenesisJoinHint, Instance: isaacoperation.SuffrageGenesisJoin{}},
	{Hint: isaacoperation.SuffrageJoinHint, Instance: isaacoperation.SuffrageJoin{}},

	{Hint: statecurrency.AccountStateValueHint, Instance: statecurrency.AccountStateValue{}},
	{Hint: statecurrency.BalanceStateValueHint, Instance: statecurrency.BalanceStateValue{}},
	{Hint: statecurrency.CurrencyDesignStateValueHint, Instance: statecurrency.CurrencyDesignStateValue{}},

	{Hint: stateextension.ContractAccountStateValueHint, Instance: stateextension.ContractAccountStateValue{}},

	{Hint: digest.AccountValueHint, Instance: digest.AccountValue{}},
	{Hint: digest.OperationValueHint, Instance: digest.OperationValue{}},
	{Hint: digestisaac.ManifestHint, Instance: digestisaac.Manifest{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: currency.CreateAccountsFactHint, Instance: currency.CreateAccountsFact{}},
	{Hint: currency.CurrencyPolicyUpdaterFactHint, Instance: currency.CurrencyPolicyUpdaterFact{}},
	{Hint: currency.CurrencyRegisterFactHint, Instance: currency.CurrencyRegisterFact{}},
	{Hint: currency.KeyUpdaterFactHint, Instance: currency.KeyUpdaterFact{}},
	{Hint: currency.SuffrageInflationFactHint, Instance: currency.SuffrageInflationFact{}},
	{Hint: currency.TransfersFactHint, Instance: currency.TransfersFact{}},

	{Hint: extension.CreateContractAccountsFactHint, Instance: extension.CreateContractAccountsFact{}},
	{Hint: extension.WithdrawsFactHint, Instance: extension.WithdrawsFact{}},

	{Hint: isaacoperation.GenesisNetworkPolicyFactHint, Instance: isaacoperation.GenesisNetworkPolicyFact{}},
	{Hint: isaacoperation.SuffrageCandidateFactHint, Instance: isaacoperation.SuffrageCandidateFact{}},
	{Hint: isaacoperation.SuffrageDisjoinFactHint, Instance: isaacoperation.SuffrageDisjoinFact{}},
	{Hint: isaacoperation.SuffrageGenesisJoinFactHint, Instance: isaacoperation.SuffrageGenesisJoinFact{}},
	{Hint: isaacoperation.SuffrageJoinFactHint, Instance: isaacoperation.SuffrageJoinFact{}},
}

func init() {
	Hinters = make([]encoder.DecodeDetail, len(launch.Hinters)+len(AddedHinters))
	copy(Hinters, launch.Hinters)
	copy(Hinters[len(launch.Hinters):], AddedHinters)

	SupportedProposalOperationFactHinters = make(
		[]encoder.DecodeDetail,
		len(launch.SupportedProposalOperationFactHinters)+len(AddedSupportedHinters),
	)
	copy(SupportedProposalOperationFactHinters, launch.SupportedProposalOperationFactHinters)
	copy(SupportedProposalOperationFactHinters[len(launch.SupportedProposalOperationFactHinters):],
		AddedSupportedHinters,
	)
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
