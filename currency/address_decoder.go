package currency

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

// func (ad *AddressDecoder) Encode(enc encoder.Encoder) (base.Address, error) {
// 	var target base.Address

// 	switch i, err := enc.DecodeWithHint(ad.b, ad.ht); {
// 	case err != nil:
// 		return nil, err
// 	case i == nil:
// 		return nil, nil
// 	default:
// 		err = util.InterfaceSetValue(i, target)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	return target, nil
// }
