package types

import (
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (ca Address) MarshalText() ([]byte, error) {
	return ca.Bytes(), nil
}

func (ca *Address) DecodeJSON(b []byte, _ encoder.Encoder) error {
	*ca = NewAddress(string(b))

	return nil
}
