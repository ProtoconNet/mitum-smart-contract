package processor

import (
	"github.com/imfact-labs/currency-model/operation/currency"
	"github.com/imfact-labs/currency-model/operation/extension"
	"github.com/imfact-labs/currency-model/operation/extras"
	cprocessor "github.com/imfact-labs/currency-model/operation/processor"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/operation/contract"

	"github.com/pkg/errors"
)

const (
	DuplicationTypeSender   ctypes.DuplicationKeyType = "sender"
	DuplicationTypeCurrency ctypes.DuplicationKeyType = "currency"
	DuplicationTypeContract ctypes.DuplicationKeyType = "contract"
)

func DuplicationKey(key string, duplType ctypes.DuplicationKeyType) string {
	return cprocessor.DuplicationKey(duplType, key)
}

func CheckDuplication(opr *cprocessor.OperationProcessor, op base.Operation) error {
	opr.Lock()
	defer opr.Unlock()

	var duplicationTypeSenderID string
	var duplicationTypeCurrencyID string
	var duplicationTypeContractID string
	var newAddresses []base.Address

	switch t := op.(type) {
	case currency.CreateAccount:
		fact, ok := t.Fact().(currency.CreateAccountFact)
		if !ok {
			return errors.Errorf("expected CreateAccountFact, not %T", t.Fact())
		}
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.UpdateKey:
		fact, ok := t.Fact().(currency.UpdateKeyFact)
		if !ok {
			return errors.Errorf("expected UpdateKeyFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.Transfer:
		fact, ok := t.Fact().(currency.TransferFact)
		if !ok {
			return errors.Errorf("expected TransferFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case currency.RegisterCurrency:
		fact, ok := t.Fact().(currency.RegisterCurrencyFact)
		if !ok {
			return errors.Errorf("expected RegisterCurrencyFact, not %T", t.Fact())
		}
		duplicationTypeCurrencyID = DuplicationKey(fact.Currency().Currency().String(), DuplicationTypeCurrency)
	case currency.UpdateCurrency:
		fact, ok := t.Fact().(currency.UpdateCurrencyFact)
		if !ok {
			return errors.Errorf("expected UpdateCurrencyFact, not %T", t.Fact())
		}
		duplicationTypeCurrencyID = DuplicationKey(fact.Currency().String(), DuplicationTypeCurrency)
	case currency.Mint:
	case extension.CreateContractAccount:
		fact, ok := t.Fact().(extension.CreateContractAccountFact)
		if !ok {
			return errors.Errorf("expected CreateContractAccountFact, not %T", t.Fact())
		}
		as, err := fact.Targets()
		if err != nil {
			return errors.Errorf("failed to get Addresses")
		}
		newAddresses = as
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeContractID = DuplicationKey(fact.Sender().String(), DuplicationTypeContract)
	case extension.Withdraw:
		fact, ok := t.Fact().(extension.WithdrawFact)
		if !ok {
			return errors.Errorf("expected WithdrawFact, not %T", t.Fact())
		}
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
	case contract.RegisterContract:
		fact, ok := t.Fact().(contract.RegisterContractFact)
		if !ok {
			return errors.Errorf("expected %T, not %T", contract.RegisterContractFact{}, t.Fact())
		}
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeContractID = DuplicationKey(fact.Contract().String(), DuplicationTypeContract)
	case contract.CallContract:
		fact, ok := t.Fact().(contract.CallContractFact)
		if !ok {
			return errors.Errorf("expected %T, not %T", contract.CallContractFact{}, t.Fact())
		}
		duplicationTypeSenderID = DuplicationKey(fact.Sender().String(), DuplicationTypeSender)
		duplicationTypeContractID = DuplicationKey(fact.Contract().String(), DuplicationTypeContract)
	default:
		return nil
	}

	if len(duplicationTypeSenderID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeSenderID]; found {
			return errors.Errorf("proposal cannot have duplicated sender, %v", duplicationTypeSenderID)
		}

		opr.Duplicated[duplicationTypeSenderID] = struct{}{}
	}

	if len(duplicationTypeCurrencyID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeCurrencyID]; found {
			return errors.Errorf(
				"cannot register duplicated currency id, %v within a proposal",
				duplicationTypeCurrencyID,
			)
		}

		opr.Duplicated[duplicationTypeCurrencyID] = struct{}{}
	}
	if len(duplicationTypeContractID) > 0 {
		if _, found := opr.Duplicated[duplicationTypeContractID]; found {
			return errors.Errorf(
				"cannot use a duplicated contract, %v within a proposal",
				duplicationTypeContractID,
			)
		}

		opr.Duplicated[duplicationTypeContractID] = struct{}{}
	}

	if len(newAddresses) > 0 {
		for i := range newAddresses {
			key := cprocessor.DuplicationKey(extras.DuplicationKeyTypeNewAddress, newAddresses[i].String())
			if _, found := opr.Duplicated[key]; found {
				return errors.Errorf("cannot use a duplicated new address, %v within a proposal", key)
			}

			opr.Duplicated[key] = struct{}{}
		}
	}

	return nil
}

func GetNewProcessor(opr *cprocessor.OperationProcessor, op base.Operation) (base.OperationProcessor, bool, error) {
	switch i, err := opr.GetNewProcessorFromHintset(op); {
	case err != nil:
		return nil, false, err
	case i != nil:
		return i, true, nil
	}

	switch t := op.(type) {
	case currency.CreateAccount,
		currency.UpdateKey,
		currency.Transfer,
		currency.RegisterCurrency,
		currency.UpdateCurrency,
		currency.Mint,
		extension.CreateContractAccount,
		extension.UpdateHandler,
		extension.Withdraw,
		contract.RegisterContract,
		contract.CallContract:
		return nil, false, errors.Errorf("%T needs SetProcessor", t)
	default:
		return nil, false, nil
	}
}
