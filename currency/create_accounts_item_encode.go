package currency

import (
	"fmt"

	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (it *BaseCreateAccountsItem) unpack(enc encoder.Encoder, ht hint.Hint, bks []byte, bam []byte, sadtype string) error {
	e := util.StringErrorFunc("failed to unmarshal BaseCreateAccountsItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	if hinter, err := enc.Decode(bks); err != nil {
		return e(err, "")
	} else if k, ok := hinter.(AccountKeys); !ok {
		return util.ErrWrongType.Errorf("expected AccountsKeys, not %T", hinter)
	} else {
		it.keys = k
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e(err, "")
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	it.amounts = amounts
	it.addressType = hint.Type(sadtype)
	fmt.Println(it.addressType)

	return nil
}
