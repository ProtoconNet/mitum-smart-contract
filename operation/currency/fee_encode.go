package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *FeeOperationFact) unpack(
	enc encoder.Encoder,
	bam []byte,
) error {
	e := util.StringErrorFunc("failed to unmarshal FeeOperationFact")

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e(err, "")
	}

	amounts := make([]types.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(types.Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	fact.amounts = amounts

	return nil
}
