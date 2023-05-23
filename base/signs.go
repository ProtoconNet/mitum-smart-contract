package base

import (
	"github.com/ProtoconNet/mitum2/base"
)

func CheckFactSignsByPubs(pubs []base.Publickey, threshold base.Threshold, signs []base.Sign) error {
	var signed uint
	for i := range signs {
		for j := range pubs {
			if signs[i].Signer().Equal(pubs[j]) {
				signed++

				break
			}
		}
	}

	if float64(signed) < threshold.Float64() {
		return base.NewBaseOperationProcessReasonError("not enough suffrage signs")
	}

	return nil
}
