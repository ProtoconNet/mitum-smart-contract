package common

import (
	"fmt"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var (
	ErrDecodeJson           = util.NewIDError(string(ErrMDecodeJson))
	ErrDecodeBson           = util.NewIDError(string(ErrMDecodeBson))
	ErrFactInvalid          = util.NewIDError(string(ErrMFactInvalid))
	ErrItemInvalid          = util.NewIDError(string(ErrMItemInvalid))
	ErrNodeOperationInvalid = util.NewIDError(string(ErrMNodeOperationInvalid))
	ErrOperationInvalid     = util.NewIDError(string(ErrMOperationInvalid))
	ErrPreProcess           = util.NewIDError(string(ErrMPreProcess))
	ErrStateInvalid         = util.NewIDError(string(ErrMStateInvalid))
)

var (
	ErrAccountE        = util.NewIDError(string(ErrMAccountE))
	ErrAccountNAth     = util.NewIDError(string(ErrMAccountNAth))
	ErrAccountNF       = util.NewIDError(string(ErrMAccountNF))
	ErrAccTypeInvalid  = util.NewIDError(string(ErrMAccTypeInvalid))
	ErrArrayLen        = util.NewIDError(string(ErrMArrayLen))
	ErrCAccountE       = util.NewIDError(string(ErrMCAccountE))
	ErrCAccountNA      = util.NewIDError(string(ErrMCAccountNA))
	ErrCAccountNF      = util.NewIDError(string(ErrMCAccountNF))
	ErrCurrencyNF      = util.NewIDError(string(ErrMCurrencyNF))
	ErrDupVal          = util.NewIDError(string(ErrMDupVal))
	ErrSelfTarget      = util.NewIDError(string(ErrMSelfTarget))
	ErrServiceE        = util.NewIDError(string(ErrMServiceE))
	ErrServiceNF       = util.NewIDError(string(ErrMServiceNF))
	ErrSignInvalid     = util.NewIDError(string(ErrMSignInvalid))
	ErrSignNE          = util.NewIDError(string(ErrMSignNE))
	ErrStateE          = util.NewIDError(string(ErrMStateE))
	ErrStateNF         = util.NewIDError(string(ErrMStateNF))
	ErrStateValInvalid = util.NewIDError(string(ErrMStateValInvalid))
	ErrTypeMismatch    = util.NewIDError(string(ErrMTypeMismatch))
	ErrValOOR          = util.NewIDError(string(ErrMValOOR))
	ErrValueInvalid    = util.NewIDError(string(ErrMValueInvalid))
)

var (
	ErrMDecodeJson           = ErrMessage("Decode Json")
	ErrMDecodeBson           = ErrMessage("Decode Bson")
	ErrMFactInvalid          = ErrMessage("Invalid fact")
	ErrMItemInvalid          = ErrMessage("Invalid item")
	ErrMNodeOperationInvalid = ErrMessage("Invalid BaseNodeOperation")
	ErrMOperationInvalid     = ErrMessage("Invalid BaseOperation")
	ErrMPreProcess           = ErrMessage("PreProcess")
	ErrMStateInvalid         = ErrMessage("Invalid BaseState")
)

var (
	ErrMAccountE        = ErrMessage("Account exist")
	ErrMAccountNAth     = ErrMessage("Account not authorized")
	ErrMAccountNF       = ErrMessage("Account not found")
	ErrMAccTypeInvalid  = ErrMessage("Invalid account type")
	ErrMArrayLen        = ErrMessage("Array length")
	ErrMCAccountE       = ErrMessage("Contract account exist")
	ErrMCAccountNA      = ErrMessage("Contract account not allowed")
	ErrMCAccountNF      = ErrMessage("Contract account not found")
	ErrMCurrencyE       = ErrMessage("Currency exist")
	ErrMCurrencyNF      = ErrMessage("Currency not found")
	ErrMDupVal          = ErrMessage("Duplicated value")
	ErrMSignInvalid     = ErrMessage("Invalid signing")
	ErrMSignNE          = ErrMessage("Not enough sign")
	ErrMSelfTarget      = ErrMessage("Self targeted")
	ErrMServiceE        = ErrMessage("Service exist")
	ErrMServiceNF       = ErrMessage("Service not found")
	ErrMStateE          = ErrMessage("State exist")
	ErrMStateNF         = ErrMessage("State not found")
	ErrMStateValInvalid = ErrMessage("Invalid state value")
	ErrMTypeMismatch    = ErrMessage("Type mismatch")
	ErrMValOOR          = ErrMessage("Value out of range")
	ErrMValueInvalid    = ErrMessage("Invalid value")
)

type ErrMessage string

func (e ErrMessage) Wrap(s ErrMessage) ErrMessage {
	return ErrMessage(fmt.Sprintf("%s: %s", e, s))
}

func (e ErrMessage) Errorf(format string, args ...interface{}) string {
	return fmt.Sprintf("%s: "+format, append([]interface{}{e}, args...)...)
}

func DecorateError(err, v error, arg interface{}) error {
	nerr := err

	switch {
	case errors.Is(v, ErrDecodeBson):
		if !errors.Is(err, ErrDecodeBson) {
			nerr = ErrDecodeBson.Wrap(errors.Errorf("%T: %v", arg, err))
		}
	case errors.Is(v, ErrDecodeJson):
		if !errors.Is(err, ErrDecodeJson) {
			nerr = ErrDecodeJson.Wrap(errors.Errorf("%T: %v", arg, err))
		}
	case errors.Is(v, ErrFactInvalid):
		if !errors.Is(err, ErrFactInvalid) {
			nerr = ErrFactInvalid.Wrap(errors.Errorf("%T: %v", arg, err))
		}
	case errors.Is(v, ErrItemInvalid):
		if !errors.Is(err, ErrItemInvalid) {
			nerr = ErrItemInvalid.Wrap(errors.Errorf("%T: %v", arg, err))
		}
	default:
	}

	return nerr
}
