package currency

import (
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *CurrencyRegisterFact) unpack(
	enc encoder.Encoder,
	ufact CurrencyRegisterFactJSONUnMarshaler,
) error {
	fact.BaseFact.SetJSONUnmarshaler(ufact.BaseFactJSONUnmarshaler)

	return encoder.Decode(enc, ufact.CR, &fact.currency)
}
