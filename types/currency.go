package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/pkg/errors"
	"regexp"
)

var (
	MinLengthCurrencyID = 3
	MaxLengthCurrencyID = 10
	ReCurrencyID        = `[A-Z0-9][A-Z0-9_\.\!\$\*\@]*[A-Z0-9]`
	ReValidCurrencyID   = regexp.MustCompile(`^` + ReCurrencyID + `$`)
	ReSpecialCh         = `[^\s:/?#\[\]$@]*`
	ReValidSpcecialCh   = regexp.MustCompile(`^` + ReSpecialCh + `$`)
)

type CurrencyID string

func (cid CurrencyID) Bytes() []byte {
	return []byte(cid)
}

func (cid CurrencyID) String() string {
	return string(cid)
}

func (cid CurrencyID) IsValid([]byte) error {
	if l := len(cid); l < MinLengthCurrencyID || l > MaxLengthCurrencyID {
		return common.ErrValOOR.Wrap(
			errors.Errorf("invalid length of currency id, %d <= %d <= %d", MinLengthCurrencyID, l, MaxLengthCurrencyID))
	} else if !ReValidCurrencyID.Match([]byte(cid)) {
		return common.ErrValueInvalid.Wrap(
			errors.Errorf("currency ID %v, must match regex `^[A-Z0-9][A-Z0-9_\\.\\!\\$\\*\\@]*[A-Z0-9]$`", cid))
	}

	return nil
}
