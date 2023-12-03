package types

import (
	"regexp"

	"github.com/ProtoconNet/mitum2/util"
)

var (
	MinLengthCurrencyID = 3
	MaxLengthCurrencyID = 10
	ReValidCurrencyID   = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_\.\!\$\*\@]*[A-Z0-9]$`)
	ReSpcecialChar      = regexp.MustCompile(`^[^\s:/?#\[\]@]*$`)
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
		return util.ErrInvalid.Errorf(
			"invalid length of currency id, %d <= %d <= %d", MinLengthCurrencyID, l, MaxLengthCurrencyID)
	} else if !ReValidCurrencyID.Match([]byte(cid)) {
		return util.ErrInvalid.Errorf("wrong currency id, %v", cid)
	}

	return nil
}
