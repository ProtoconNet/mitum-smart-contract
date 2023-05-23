package base

import (
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

func (ca Address) MarshalText() ([]byte, error) {
	return ca.Bytes(), nil
}

func (ca *Address) DecodeJSON(b []byte, _ *jsonenc.Encoder) error {
	*ca = NewAddress(string(b))

	return nil
}

func (ca EthAddress) MarshalText() ([]byte, error) {
	return ca.Bytes(), nil
}

func (ca *EthAddress) DecodeJSON(b []byte, _ *jsonenc.Encoder) error {
	*ca = NewEthAddress(string(b))

	return nil
}
