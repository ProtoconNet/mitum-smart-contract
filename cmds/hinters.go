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
	{Hint: types.EthAccountKeysHint, Instance: types.EthAccountKeys{}},
	{Hint: types.AddressHint, Instance: types.Address{}},
	{Hint: types.AmountHint, Instance: types.Amount{}},
	{Hint: types.ContractAccountKeysHint, Instance: types.ContractAccountKeys{}},
	{Hint: types.ContractAccountStatusHint, Instance: types.ContractAccountStatus{}},
	{Hint: types.CurrencyDesignHint, Instance: types.CurrencyDesign{}},
	{Hint: types.CurrencyPolicyHint, Instance: types.CurrencyPolicy{}},
	{Hint: types.EthAddressHint, Instance: types.EthAddress{}},
	{Hint: types.FixedFeeerHint, Instance: types.FixedFeeer{}},
	{Hint: types.MEPrivatekeyHint, Instance: types.MEPrivatekey{}},
	{Hint: types.MEPublickeyHint, Instance: types.MEPublickey{}},
	{Hint: types.NilFeeerHint, Instance: types.NilFeeer{}},
	{Hint: types.RatioFeeerHint, Instance: types.RatioFeeer{}},

	{Hint: currency.CreateAccountHint, Instance: currency.CreateAccount{}},
	{Hint: currency.CreateAccountItemMultiAmountsHint, Instance: currency.CreateAccountItemMultiAmounts{}},
	{Hint: currency.CreateAccountItemSingleAmountHint, Instance: currency.CreateAccountItemSingleAmount{}},
	{Hint: currency.UpdateCurrencyHint, Instance: currency.UpdateCurrency{}},
	{Hint: currency.RegisterCurrencyHint, Instance: currency.RegisterCurrency{}},
	//{Hint: currency.FeeOperationFactHint, Instance: currency.FeeOperationFact{}},
	//{Hint: currency.FeeOperationHint, Instance: currency.FeeOperation{}},
	{Hint: currency.RegisterGenesisCurrencyHint, Instance: currency.RegisterGenesisCurrency{}},
	{Hint: currency.RegisterGenesisCurrencyFactHint, Instance: currency.RegisterGenesisCurrencyFact{}},
	{Hint: currency.UpdateKeyHint, Instance: currency.UpdateKey{}},
	{Hint: currency.MintHint, Instance: currency.Mint{}},
	{Hint: currency.TransferHint, Instance: currency.Transfer{}},
	{Hint: currency.TransferItemMultiAmountsHint, Instance: currency.TransferItemMultiAmounts{}},
	{Hint: currency.TransferItemSingleAmountHint, Instance: currency.TransferItemSingleAmount{}},

	{Hint: extension.CreateContractAccountHint, Instance: extension.CreateContractAccount{}},
	{Hint: extension.CreateContractAccountItemMultiAmountsHint, Instance: extension.CreateContractAccountItemMultiAmounts{}},
	{Hint: extension.CreateContractAccountItemSingleAmountHint, Instance: extension.CreateContractAccountItemSingleAmount{}},
	{Hint: extension.UpdateOperatorHint, Instance: extension.UpdateOperator{}},
	{Hint: extension.WithdrawHint, Instance: extension.Withdraw{}},
	{Hint: extension.WithdrawItemMultiAmountsHint, Instance: extension.WithdrawItemMultiAmounts{}},
	{Hint: extension.WithdrawItemSingleAmountHint, Instance: extension.WithdrawItemSingleAmount{}},

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
	{Hint: currency.CreateAccountFactHint, Instance: currency.CreateAccountFact{}},
	{Hint: currency.UpdateCurrencyFactHint, Instance: currency.UpdateCurrencyFact{}},
	{Hint: currency.RegisterCurrencyFactHint, Instance: currency.RegisterCurrencyFact{}},
	{Hint: currency.UpdateKeyFactHint, Instance: currency.UpdateKeyFact{}},
	{Hint: currency.MintFactHint, Instance: currency.MintFact{}},
	{Hint: currency.TransferFactHint, Instance: currency.TransferFact{}},

	{Hint: extension.CreateContractAccountFactHint, Instance: extension.CreateContractAccountFact{}},
	{Hint: extension.UpdateOperatorFactHint, Instance: extension.UpdateOperatorFact{}},
	{Hint: extension.WithdrawFactHint, Instance: extension.WithdrawFact{}},

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
			return errors.Wrap(err, "add to encoder")
		}
	}

	for i := range SupportedProposalOperationFactHinters {
		if err := enc.Add(SupportedProposalOperationFactHinters[i]); err != nil {
			return errors.Wrap(err, "add to encoder")
		}
	}

	return nil
}
