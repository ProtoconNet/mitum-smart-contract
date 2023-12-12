package currency

//
//import (
//	"github.com/ProtoconNet/mitum-currency/v3/types"
//	"github.com/ProtoconNet/mitum2/util"
//	"github.com/ProtoconNet/mitum2/util/encoder"
//	"github.com/pkg/errors"
//)
//
//func (fact *FeeOperationFact) unpack(
//	enc encoder.Encoder,
//	bam []byte,
//) error {
//	e := util.StringError("unmarshal FeeOperationFact")
//
//	ham, err := enc.DecodeSlice(bam)
//	if err != nil {
//		return e.Wrap(err)
//	}
//
//	amounts := make([]types.Amount, len(ham))
//	for i := range ham {
//		j, ok := ham[i].(types.Amount)
//		if !ok {
//			return errors.Errorf("expected Amount, not %T", ham[i])
//		}
//
//		amounts[i] = j
//	}
//
//	fact.amounts = amounts
//
//	return nil
//}
