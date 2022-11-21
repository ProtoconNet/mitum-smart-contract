package currency

import (
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *FeeOperationFact) unpack(
	enc encoder.Encoder,
	ufact FeeOperationFactJSONUnMarshaler,
) error {
	fact.BaseFact.SetJSONUnmarshaler(ufact.BaseFactJSONUnmarshaler)

	ham, err := enc.DecodeSlice(ufact.AM)
	if err != nil {
		return err
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	fact.amounts = amounts

	return nil
}
