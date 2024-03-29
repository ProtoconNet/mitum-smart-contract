package common

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

func IsValidSignFact(sf base.SignFact, networkID []byte) error {
	sfs := sf.Signs()
	if len(sfs) < 1 {
		return ErrSignInvalid.Wrap(errors.Errorf("empty signs"))
	}

	bs := make([]util.IsValider, len(sf.Signs())+1)
	bs[0] = sf.Fact()

	for i := range sfs {
		bs[i+1] = sfs[i]
	}

	if err := util.CheckIsValiders(networkID, false, bs...); err != nil {
		return err
	}

	// NOTE caller should check the duplication of Signs

	for i := range sfs {
		if err := sfs[i].Verify(networkID, sf.Fact().Hash().Bytes()); err != nil {
			return ErrSignInvalid.Wrap(errors.Errorf("verify sign: %v", err))
		}
	}

	return nil
}
