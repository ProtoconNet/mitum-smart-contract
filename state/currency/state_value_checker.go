package currency

//
//import (
//	"context"
//	"github.com/ProtoconNet/mitum2/base"
//	"github.com/pkg/errors"
//)
//
//type CheckerFunc func(context.Context, base.GetStateFunc) (bool, error)
//
//type Middleware func(CheckerFunc) CheckerFunc
//
//func ExistAccount(next CheckerFunc) CheckerFunc {
//	// Account가 존재하는 지 확인
//	return func(ctx context.Context, getStateFunc base.GetStateFunc) (bool, error) {
//
//		_, found, err := getStateFunc(st.Key())
//		if !found {
//			return false, errors.Errorf("Account not found")
//		} else if err != nil {
//			return false, err
//		}
//		return next(ctx, getStateFunc)
//	}
//}
//
//func ExistContractAccount(next CheckerFunc) CheckerFunc {
//	// ContractAccount가 존재하는 지 확인
//	return f
//}
//
//func IsSingleSigAccount(next CheckerFunc) CheckerFunc {
//	// SingleSigAccount인지 확인
//	return f
//}
//
//func IsContractAccountOwner(next CheckerFunc) CheckerFunc {
//	// SingleSigAccount인지 확인
//	return f
//}
