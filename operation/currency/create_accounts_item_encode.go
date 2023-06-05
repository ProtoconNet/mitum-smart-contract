package currency

import (
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *BaseCreateAccountsItem) unpack(enc encoder.Encoder, ht hint.Hint, bks []byte, bam []byte, sadtype string) error {
	e := util.StringErrorFunc("failed to unmarshal BaseCreateAccountsItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	if hinter, err := enc.Decode(bks); err != nil {
		return e(err, "")
	} else if k, ok := hinter.(base2.AccountKeys); !ok {
		return util.ErrWrongType.Errorf("expected AccountsKeys, not %T", hinter)
	} else {
		it.keys = k
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e(err, "")
	}

	amounts := make([]base2.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(base2.Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	it.amounts = amounts
	it.addressType = hint.Type(sadtype)

	return nil
}
